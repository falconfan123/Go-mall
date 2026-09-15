// Package closeorder 承载"关闭超时未支付订单"业务（design 决策 7）：
// delay consumer 与对账扫描器（经 order RPC）共用同一入口，消除内联复制。
// 本包不依赖 svc（避免 delay→logic→svc→delay 导入环），依赖全部显式注入。
package closeorder

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/common/consts/code"
	ordertypes "github.com/falconfan123/Go-mall/common/types/order"
	ordermodel "github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CloseExpiredOrder struct {
	orderModel   ordermodel.OrdersModel
	orderItems   ordermodel.OrderItemsModel
	inventoryRpc inventoryclient.Inventory
	conn         sqlx.SqlConn
	logx.Logger
}

func New(om ordermodel.OrdersModel, oim ordermodel.OrderItemsModel, inv inventoryclient.Inventory, conn sqlx.SqlConn, log logx.Logger) *CloseExpiredOrder {
	return &CloseExpiredOrder{orderModel: om, orderItems: oim, inventoryRpc: inv, conn: conn, Logger: log}
}

// Close 关闭超时未支付订单：状态机前置校验（Created 或 PendingPayment → Closed/Expired）
// + ReturnPreInventory。实施期扩展（方案 1，design 决策 7 补记）：scan2 的目标孤儿形态
// 是 PendingPayment（支付链接已开出但超时未付），其 payment 超时兜底原由休眠的
// PaymentDelayMQ 负责——delay 队列只发 Created 单，故扩展前置状态不影响 delay 语义。
//
// 幂等响应语义：其他状态（已支付/已关闭）→ Success + "already closed, skip"（重试安全）；
// ReturnPreInventory 失败 → 返回错误（调用方重试，连续失败由调用方产出 exception）。
// 订单不存在 → 返回错误（调用方重试，超限转人工——数据矛盾级别的诚实处理）。
func (c *CloseExpiredOrder) Close(ctx context.Context, orderId string, userId int32) (statusCode int32, statusMsg string, err error) {
	if orderId == "" || userId == 0 {
		return 0, "", fmt.Errorf("参数错误")
	}
	res := struct {
		Code int32
		Msg  string
		Skip bool
	}{Code: code.Success}
	var preOrderId string
	var uid int32
	if err := c.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		orderRes, err := c.orderModel.WithSession(session).GetOrderByOrderIDAndUserIDWithLock(ctx, orderId, userId)
		if err != nil {
			return err
		}
		// 超时未支付的两段孤儿形态（design 决策 7 实施期扩展）：
		//   Created(1)——delay 队列目标；PendingPayment(2)——scan2 目标（支付链接已开出）
		st := ordertypes.OrderStatus(orderRes.OrderStatus)
		if st != ordertypes.OrderStatusCreated && st != ordertypes.OrderStatusPendingPayment {
			res.Code = code.Success
			res.Msg = "already closed, skip"
			res.Skip = true
			c.Logger.Infow("order not in closable status, skip closing",
				logx.Field("order_id", orderId), logx.Field("user_id", userId),
				logx.Field("order_status", orderRes.OrderStatus))
			return nil
		}
		if st == ordertypes.OrderStatusPendingPayment {
			res.Msg = "closed expired pending payment"
		}
		if err := c.orderModel.WithSession(session).UpdateOrderStatusByOrderIDAndUserID(
			ctx, orderId, userId,
			ordertypes.OrderStatusClosed, ordertypes.PaymentStatusExpired); err != nil {
			return err
		}
		preOrderId = orderRes.PreOrderId
		uid = int32(orderRes.UserId)
		return nil
	}); err != nil {
		c.Logger.Errorw("failed to update order status",
			logx.Field("err", err), logx.Field("order_id", orderId), logx.Field("user_id", userId))
		return 0, "", err
	}

	if res.Skip {
		// 幂等跳过：非 Created 不触发任何副作用（含 ReturnPreInventory）
		return res.Code, res.Msg, nil
	}

	orderItems, err := c.orderItems.QueryOrderItemsByOrderID(ctx, orderId)
	if err != nil {
		c.Logger.Errorw("failed to query order items",
			logx.Field("err", err), logx.Field("order_id", orderId))
		return 0, "", err
	}
	itemsReq := make([]*inventoryclient.InventoryReq_Items, len(orderItems))
	for i, orderItem := range orderItems {
		itemsReq[i] = &inventoryclient.InventoryReq_Items{
			ProductId: int32(orderItem.ProductId),
			Quantity:  int32(orderItem.Quantity),
		}
	}
	returnPreInventoryResp, err := c.inventoryRpc.ReturnPreInventory(ctx, &inventoryclient.InventoryReq{
		PreOrderId: preOrderId,
		Items:      itemsReq,
		UserId:     uid,
	})
	if err != nil {
		c.Logger.Errorw("failed to return pre inventory",
			logx.Field("err", err), logx.Field("order_id", orderId))
		return 0, "", err
	}
	// 库存幂等兜底：已释放过的响应视为成功，仅记录提示日志
	if returnPreInventoryResp.StatusCode != code.Success {
		c.Logger.Infow("info to return pre inventory",
			logx.Field("status_msg", returnPreInventoryResp.StatusMsg), logx.Field("order_id", orderId))
	}
	return res.Code, res.Msg, nil
}
