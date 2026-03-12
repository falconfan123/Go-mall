package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/types/coupons"
	"github.com/falconfan123/Go-mall/services/coupons/internal/svc"
	couponspb "github.com/falconfan123/Go-mall/services/coupons/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReleaseCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseCouponLogic {
	return &ReleaseCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReleaseCoupon 释放优惠券（Saga补偿操作）
func (l *ReleaseCouponLogic) ReleaseCoupon(in *couponspb.ReleaseCouponReq) (*couponspb.EmptyResp, error) {

	res := &couponspb.EmptyResp{}
	// --------------- check ---------------
	if in.UserId == 0 || len(in.UserCouponId) == 0 || len(in.PreOrderId) == 0 {
		res.StatusCode = code.NotWithParam
		res.StatusMsg = code.NotWithParamMsg
		return nil, status.Error(codes.Aborted, code.NotWithParamMsg)
	}
	// --------------- 事务操作 ---------------
	if err := l.svcCtx.Model.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 检查优惠券锁定状态与订单匹配
		state, err := l.svcCtx.UserCouponsModel.CheckUserCouponStatus(l.ctx, session, uint64(in.UserId), in.UserCouponId)
		if err != nil {
			l.Logger.Errorw("check lock status failed", logx.Field("error", err))
			return err
		}

		// 2. 状态校验（幂等性保障）
		if coupons.CouponStatus(state) != coupons.CouponStatusLocked {
			l.Logger.Infow("coupon status is not locked", logx.Field("userId", in.UserId), logx.Field("couponId", in.UserCouponId))
			res.StatusCode = code.CouponStatusInvalid
			res.StatusMsg = code.CouponStatusInvalidMsg
			return nil
		}

		// 3. 执行状态更新
		if err := l.svcCtx.UserCouponsModel.UpdateStatusOrderById(
			l.ctx,
			"", // 清空ID
			int(in.UserId),
			coupons.CouponStatusAvailable,
		); err != nil {
			l.Logger.Errorw("update coupon status failed", logx.Field("error", err))
			return err
		}
		return nil
	}); err != nil {
		l.Logger.Errorw("transact release coupon error", logx.Field("err", err))
		return nil, status.Error(codes.Internal, code.ServerErrorMsg) // 错误已携带正确status
	}
	if res.StatusCode != code.Success {
		return nil, status.Error(codes.Aborted, res.StatusMsg)
	}
	return res, nil
}
