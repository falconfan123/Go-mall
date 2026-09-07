package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/types/coupons"
	"github.com/falconfan123/Go-mall/dal/model/coupons/user_coupons"
	"github.com/falconfan123/Go-mall/services/coupons/internal/svc"
	couponspb "github.com/falconfan123/Go-mall/services/coupons/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UseCouponRollbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUseCouponRollbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UseCouponRollbackLogic {
	return &UseCouponRollbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UseCouponRollback 使用优惠券回滚（Saga 补偿分支，design 决策 2）。
// Saga 补偿回调与 action 共用同类型 payload（UseCouponReq），故补偿方法接受 UseCouponReq；
// 既有 ReleaseCoupon(ReleaseCouponReq) 签名不兼容，由本方法承担 saga 补偿职责。
//
// 语义（与 specs"业务幂等兜底容忍"契约一致）：
//   - 券不存在/已释放/仍锁定（action 未提交）→ 容忍并返回成功；
//   - 已使用 → 回滚为可用并撤销使用流水；
//   - 基础设施失败 → 原样透传（dtm 退避重试）。
func (l *UseCouponRollbackLogic) UseCouponRollback(in *couponspb.UseCouponReq) (*couponspb.EmptyResp, error) {
	res := &couponspb.EmptyResp{}
	if in.UserId == 0 || in.CouponId == "" || in.OrderId == "" {
		return nil, status.Error(codes.Aborted, code.NotWithParamMsg)
	}

	if bar, berr := dtmgrpc.BarrierFromGrpc(l.ctx); berr == nil {
		// saga 补偿分支路径：barrier 自管事务（origin=action，空补偿/重复补偿短路）
		bar.DBType = "postgres"
		if err := bar.CallWithDB(l.svcCtx.DtmDB, func(tx *sql.Tx) error {
			return l.rollbackTx(l.ctx, sqlx.NewSessionFromTx(tx), in, res)
		}); err != nil {
			l.Logger.Errorw("saga use coupon rollback failed", logx.Field("err", err),
				logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId),
				logx.Field("order_id", in.OrderId))
			return nil, err
		}
		return res, nil
	}

	// 非 saga 直连路径：既有事务语义（兼容调用）
	if err := l.svcCtx.Model.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		return l.rollbackTx(ctx, session, in, res)
	}); err != nil {
		l.Logger.Errorw("transact use coupon rollback error", logx.Field("err", err))
		return nil, status.Error(codes.Internal, code.ServerErrorMsg)
	}
	return res, nil
}

func (l *UseCouponRollbackLogic) rollbackTx(ctx context.Context, session sqlx.Session, in *couponspb.UseCouponReq, res *couponspb.EmptyResp) error {
	uc, err := l.svcCtx.UserCouponsModel.GetUserCouponByUserIdCouponIdWithLock(ctx, session, uint64(in.UserId), in.CouponId)
	if err != nil {
		if errors.Is(err, user_coupons.ErrNotFound) || errors.Is(err, sqlx.ErrNotFound) {
			l.Logger.Infow("user coupon not found on rollback, tolerate", logx.Field("user_id", in.UserId),
				logx.Field("coupon_id", in.CouponId), logx.Field("order_id", in.OrderId))
			return nil
		}
		return err
	}

	switch coupons.CouponStatus(uc.Status) {
	case coupons.CouponStatusUsed:
		// 回滚为可用 + 撤销使用流水
		if err := l.svcCtx.UserCouponsModel.WithSession(session).UpdateStatusOrderById(
			ctx, "", int(uc.Id), coupons.CouponStatusAvailable); err != nil {
			l.Logger.Errorw("revert coupon status failed", logx.Field("err", err),
				logx.Field("coupon_id", in.CouponId), logx.Field("order_id", in.OrderId))
			return err
		}
		if err := l.svcCtx.CouponUsageModel.DeleteByOrderId(ctx, session, in.OrderId); err != nil {
			l.Logger.Errorw("delete coupon usage failed", logx.Field("err", err),
				logx.Field("coupon_id", in.CouponId), logx.Field("order_id", in.OrderId))
			return err
		}
		l.Logger.Infow("coupon rolled back", logx.Field("user_id", in.UserId),
			logx.Field("coupon_id", in.CouponId), logx.Field("order_id", in.OrderId))
		return nil
	default:
		// Available（已释放，幂等容忍）/ Locked（action 未提交，无需补偿）
		l.Logger.Infow("coupon rollback tolerated", logx.Field("user_id", in.UserId),
			logx.Field("coupon_id", in.CouponId), logx.Field("order_id", in.OrderId),
			logx.Field("status", uc.Status))
		res.StatusCode = code.Success
		return nil
	}
}
