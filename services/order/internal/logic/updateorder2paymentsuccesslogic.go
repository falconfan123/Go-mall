package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	"github.com/falconfan123/Go-mall/common/consts/code"
	ordertypes "github.com/falconfan123/Go-mall/common/types/order"
	ordermodel "github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/falconfan123/Go-mall/services/order/internal/svc"
	order "github.com/falconfan123/Go-mall/services/order/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrder2PaymentSuccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrder2PaymentSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrder2PaymentSuccessLogic {
	return &UpdateOrder2PaymentSuccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateOrder2PaymentSuccess 订单结算分支（Saga branch-01，design 决策 2）。
//
// 携带 dtm barrier 上下文（真实 saga 调用）时：经 barrier 幂等执行三分支判定——
// PendingPayment 正常流转 Paid；已 Paid 视为幂等重放返回成功（Aborted 会被 dtm
// 误判为分支失败触发错误补偿）；其他状态/订单不存在为确定性失败（Aborted → 补偿）。
// 无 barrier 上下文（直连 RPC，兼容调用）时：保持既有事务语义。
// 结算不再经过 MQ（notify 已退役，副作用由 Saga 分支直接编排）。
func (l *UpdateOrder2PaymentSuccessLogic) UpdateOrder2PaymentSuccess(in *order.UpdateOrder2PaymentSuccessRequest) (*order.EmptyRes, error) {
	if in.OrderId == "" || in.UserId == 0 || in.PaymentResult == nil {
		return nil, status.Error(codes.Aborted, "参数错误")
	}
	res := &order.EmptyRes{}

	if bar, berr := dtmgrpc.BarrierFromGrpc(l.ctx); berr == nil {
		// saga 分支路径：barrier 自管事务（成功提交/失败回滚），业务在其事务内执行
		bar.DBType = "postgres"
		if err := bar.CallWithDB(l.svcCtx.DtmDB, func(tx *sql.Tx) error {
			if err := l.settleOrderTx(l.ctx, l.svcCtx.OrderModel.WithSession(sqlx.NewSessionFromTx(tx)), in, res); err != nil {
				return err
			}
			if res.StatusCode != code.Success {
				// 业务规则失败（订单不存在/状态冲突）：确定性失败 → dtm 触发补偿
				return status.Error(codes.Aborted, res.StatusMsg)
			}
			return nil
		}); err != nil {
			l.Logger.Errorw("saga settle branch failed", logx.Field("order_id", in.OrderId), logx.Field("err", err))
			return nil, err
		}
		l.clearPaymentLocks(in.UserId, in.OrderId)
		return res, nil
	}

	// 非 saga 直连路径：既有事务语义
	if err := l.svcCtx.Model.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		return l.settleOrderTx(ctx, l.svcCtx.OrderModel.WithSession(session), in, res)
	}); err != nil {
		l.Logger.Errorw("update order status error", logx.Field("order_id", in.OrderId), logx.Field("err", err))
		return nil, status.Error(codes.Internal, "更新订单状态失败")
	}
	l.clearPaymentLocks(in.UserId, in.OrderId)
	if res.StatusCode != code.Success {
		l.Logger.Infow("UpdateOrder2PaymentSuccess process aborted",
			logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
		return nil, status.Error(codes.Aborted, res.StatusMsg)
	}
	return res, nil
}

// settleOrderTx 在给定事务会话内执行订单状态机三分支判定（design 决策 2）。
// 返回 error 仅代表基础设施失败（dtm 将退避重试）；业务规则失败通过 res.StatusCode
// 表达，由调用方转换为 codes.Aborted（确定性失败 → 补偿）。
func (l *UpdateOrder2PaymentSuccessLogic) settleOrderTx(ctx context.Context, om ordermodel.OrdersModel, in *order.UpdateOrder2PaymentSuccessRequest, res *order.EmptyRes) error {
	orderRes, err := om.GetOrderByOrderIDAndUserIDWithLock(ctx, in.OrderId, in.UserId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			res.StatusCode = code.OrderNotExist
			res.StatusMsg = code.OrderNotExistMsg
			l.Logger.Infow("order not found", logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
			return nil
		}
		l.Logger.Errorw("update order status error", logx.Field("err", err),
			logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
		return err
	}

	switch st := ordertypes.OrderStatus(orderRes.OrderStatus); {
	case st == ordertypes.OrderStatusPaid:
		// 幂等重放：已 Paid 必须返回成功（返回 Aborted 会被 dtm 误判分支失败触发补偿）
		l.Logger.Infow("order already paid, replay treated as success",
			logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
		res.StatusCode = code.Success
		return nil
	case st != ordertypes.OrderStatusPendingPayment:
		// 真实状态冲突（Closed/Cancelled 等）：确定性失败
		res.StatusCode = code.OrderStatusInvalid
		res.StatusMsg = code.OrderStatusInvalidMsg
		l.Logger.Infow("order status error", logx.Field("order_id", in.OrderId),
			logx.Field("user_id", in.UserId), logx.Field("order_status", orderRes.OrderStatus))
		return nil
	}
	// 已知悬挂风险（design 决策 2）：订单行其间被删除时补偿亦会失败，依赖
	// barrier 主保险 + dtm 补偿重试告警承接
	if ordertypes.PaymentStatus(orderRes.PaymentStatus) != ordertypes.PaymentStatusPaying {
		res.StatusCode = code.PaymentStatusInvalid
		res.StatusMsg = code.PaymentStatusInvalidMsg
		l.Logger.Infow("payment status error", logx.Field("order_id", in.OrderId),
			logx.Field("user_id", in.UserId), logx.Field("payment_status", orderRes.PaymentStatus))
		return nil
	}
	return om.UpdateOrder2Payment(ctx, in.OrderId, in.UserId, &ordertypes.PaymentResult{
		TransactionId: in.PaymentResult.TransactionId,
		PaidAmount:    in.PaymentResult.PaidAmount,
		PaidAt:        in.PaymentResult.PaidAt,
	}, ordertypes.OrderStatusPaid, ordertypes.PaymentStatusPaid)
}

// clearPaymentLocks 支付成功后清理预订单锁与库存扣减锁（幂等 DEL）。
func (l *UpdateOrder2PaymentSuccessLogic) clearPaymentLocks(userID int32, orderID string) {
	if l.svcCtx.RedisClient == nil {
		return
	}
	cacheKey := "checkout:preorder:" + fmt.Sprint(userID)
	if _, err := l.svcCtx.RedisClient.Del(cacheKey); err != nil {
		l.Logger.Errorw("delete preorder lock failed", logx.Field("err", err), logx.Field("user_id", userID))
	}
	inventoryLockKey := "inventory:deduct:" + fmt.Sprint(userID) + ":" + orderID
	if _, err := l.svcCtx.RedisClient.Del(inventoryLockKey); err != nil {
		l.Logger.Errorw("delete inventory deduct lock failed", logx.Field("err", err),
			logx.Field("user_id", userID), logx.Field("order_id", orderID))
	}
}
