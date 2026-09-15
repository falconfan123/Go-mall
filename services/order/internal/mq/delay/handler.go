package delay

import (
	"context"
	"encoding/json"

	ordermodel "github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/order/internal/closeorder"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// delayHandler 是 order-delay（死信延迟队列）的业务回调：
// 超时未支付订单关闭 -> 释放预扣库存。fall-through 与空引用风险由骨架消除。
type delayHandler struct {
	Model           sqlx.SqlConn
	OrderModel      ordermodel.OrdersModel
	OrderItemsModel ordermodel.OrderItemsModel
	InventoryRpc    inventoryclient.Inventory
}

func newDelayHandler(mq *OrderDelayMQ) *delayHandler {
	return &delayHandler{
		Model:           mq.Model,
		OrderModel:      mq.OrderModel,
		OrderItemsModel: mq.OrderItemsModel,
		InventoryRpc:    mq.InventoryRpc,
	}
}

// Decode 反序列化消息体；失败由骨架按毒消息一次即弃。
func (h *delayHandler) Decode(body []byte) (*OrderReq, error) {
	msg := &OrderReq{}
	if err := json.Unmarshal(body, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// IdempotencyKey 返回订单号作为幂等与重试计数标识。
func (h *delayHandler) IdempotencyKey(msg *OrderReq) string {
	return msg.OrderId
}

// Handle 关闭超时未支付订单并释放预扣库存——业务迁移至
// logic.CloseExpiredOrder（与对账扫描器 scan2 共用同一入口，design 决策 7），
// 本回调只做错误桥接：logic 返回错误 → 骨架按可重试失败处理；跳过语义（非 Created）
// 由 logic 以 Success+skip 表达，此处返回 nil 由骨架确认。
func (h *delayHandler) Handle(ctx context.Context, msg *OrderReq) error {
	biz := closeorder.New(h.OrderModel, h.OrderItemsModel, h.InventoryRpc, h.Model, logx.WithContext(ctx))
	if _, _, err := biz.Close(ctx, msg.OrderId, int32(msg.UserID)); err != nil {
		return err
	}
	return nil
}
