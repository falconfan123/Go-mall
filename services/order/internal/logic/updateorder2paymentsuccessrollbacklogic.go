package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	"github.com/falconfan123/Go-mall/common/consts/code"
	ordertypes "github.com/falconfan123/Go-mall/common/types/order"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/falconfan123/Go-mall/services/order/internal/svc"
	order "github.com/falconfan123/Go-mall/services/order/pb"
)

type UpdateOrder2PaymentSuccessRollbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrder2PaymentSuccessRollbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrder2PaymentSuccessRollbackLogic {
	return &UpdateOrder2PaymentSuccessRollbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateOrder2PaymentSuccessRollback 订单结算补偿分支（Saga branch-01 compensation）。
//
// 携带 dtm barrier 上下文时经 barrier 幂等执行（防补偿重放）。补偿语义与 action 相反：
// 补偿必须始终成功——订单不存在/非 Paid（action 未生效）均为容忍成功（无副作用可回滚），
// 仅基础设施失败返回错误（dtm 退避重试补偿）。实施期修复既有缺陷：原实现 DB 错误后
// 缺 return 导致 orderRes 空指针 panic。
func (l *UpdateOrder2PaymentSuccessRollbackLogic) UpdateOrder2PaymentSuccessRollback(in *order.UpdateOrder2PaymentSuccessRequest) (*order.EmptyRes, error) {
	// 参数校验保持与支付成功逻辑一致
	if in.OrderId == "" || in.UserId == 0 {
		return nil, status.Error(codes.Aborted, "参数错误")
	}
	res := &order.EmptyRes{}

	if bar, berr := dtmgrpc.BarrierFromGrpc(l.ctx); berr == nil {
		bar.DBType = "postgres"
		if err := bar.CallWithDB(l.svcCtx.DtmDB, func(tx *sql.Tx) error {
			return l.rollbackTx(l.ctx, sqlx.NewSessionFromTx(tx), in, res)
		}); err != nil {
			l.Logger.Errorw("saga order rollback failed", logx.Field("order_id", in.OrderId), logx.Field("err", err))
			return nil, err
		}
		// 补偿执行属需人工介入级别事件（specs 失败可观测契约）
		l.Logger.Errorw("saga compensation executed",
			logx.Field("branch", "order"), logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
		return res, nil
	}

	// 非 saga 直连路径：既有事务语义
	if err := l.svcCtx.Model.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		return l.rollbackTx(ctx, session, in, res)
	}); err != nil {
		l.Logger.Errorw("update order status error", logx.Field("err", err),
			logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
		return nil, status.Error(codes.Internal, "更新订单状态失败")
	}
	if res.StatusCode != code.Success {
		l.Logger.Infow("UpdateOrder2PaymentSuccessRollback process aborted",
			logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
		return nil, status.Error(codes.Aborted, res.StatusMsg)
	}
	return &order.EmptyRes{}, nil
}

// rollbackTx 在给定事务会话内执行订单补偿（barrier 路径与直连路径共用）。
// 补偿容忍语义：订单不存在/状态不满足（action 未生效）→ 成功；仅基础设施失败返回 error。
func (l *UpdateOrder2PaymentSuccessRollbackLogic) rollbackTx(ctx context.Context, session sqlx.Session, in *order.UpdateOrder2PaymentSuccessRequest, res *order.EmptyRes) error {
	// 1. 查询订单并加锁
	orderRes, err := l.svcCtx.OrderModel.WithSession(session).GetOrderByOrderIDAndUserIDWithLock(ctx, in.OrderId, in.UserId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			// 补偿容忍：无订单即无副作用可回滚（不 Aborted——补偿必须始终成功）
			res.StatusCode = code.Success
			l.Logger.Infow("order not found on rollback, tolerate as success",
				logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
			return nil
		}
		l.Logger.Errorw("query order status failed", logx.Field("err", err),
			logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
		return err
	}

	// 2. 状态检查：非 Paid（action 未生效）→ 无可回滚副作用，容忍成功
	if ordertypes.OrderStatus(orderRes.OrderStatus) != ordertypes.OrderStatusPaid ||
		ordertypes.PaymentStatus(orderRes.PaymentStatus) != ordertypes.PaymentStatusPaid {
		res.StatusCode = code.Success
		l.Logger.Infow("order not in paid state, rollback tolerated",
			logx.Field("order_status", orderRes.OrderStatus),
			logx.Field("payment_status", orderRes.PaymentStatus),
			logx.Field("user_id", in.UserId),
			logx.Field("order_id", in.OrderId))
		return nil
	}

	// 3. 更新订单状态为待支付，支付状态为失败
	if err := l.svcCtx.OrderModel.WithSession(session).UpdateOrder2PaymentRollback(
		ctx, in.OrderId, in.UserId,
	); err != nil {
		l.Logger.Errorw("update order status error", logx.Field("err", err),
			logx.Field("order_id", in.OrderId), logx.Field("user_id", in.UserId))
		return err
	}
	res.StatusCode = code.Success
	return nil
}
