package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/svc"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/coupons/coupons"
	xerrors "github.com/zeromicro/x/errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type CalculateCouponLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCalculateCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CalculateCouponLogic {
	return &CalculateCouponLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CalculateCouponLogic) CalculateCoupon(req *types.CalculateCouponReq) (resp *types.CalculateCouponResp, err error) {

	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, xerrors.New(code.AuthBlank, code.AuthBlankMsg)
	}
	calculateCouponResp, err := l.svcCtx.CouponRpc.CalculateCoupon(l.ctx, &coupons.CalculateCouponReq{
		CouponId: req.CouponID,
		Items:    convertToCouponItems(req.Items),
		UserId:   int32(userID),
	})
	if err != nil {
		logx.Errorw("call rpc CalculateCoupon failed", logx.Field("err", err))
		return nil, err
	}
	if calculateCouponResp.StatusCode != code.Success {
		return nil, xerrors.New(int(calculateCouponResp.StatusCode), calculateCouponResp.StatusMsg)
	}
	resp = &types.CalculateCouponResp{
		CouponType:     calculateCouponResp.CouponType,
		DiscountAmount: calculateCouponResp.DiscountAmount,
		FinalAmount:    calculateCouponResp.FinalAmount,
		IsUsable:       calculateCouponResp.IsUsable,
		OriginAmount:   calculateCouponResp.OriginAmount,
		UnusableReason: calculateCouponResp.UnusableReason,
	}
	return
}
