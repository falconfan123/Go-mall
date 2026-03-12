package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/svc"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/coupons/pb"
	xerrors "github.com/zeromicro/x/errors"

	"github.com/zeromicro/go-zero/core/logx"
)

// CouponRecordLogic is the business logic for couponrecord operations.
// CouponRecordLogic is the business logic for CouponRecordLogic operations.
type CouponRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCouponRecordLogic creates a new instance.
// NewCouponRecordLogic creates a new CouponRecordLogic instance.
func NewCouponRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CouponRecordLogic {
	return &CouponRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CouponRecord is a function.
//
//	does something.
func (l *CouponRecordLogic) CouponRecord(req *types.CouponListReq) (resp *types.CouponUsageListResp, err error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > biz.MaxPageSize {
		req.PageSize = biz.MaxPageSize
	}
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, xerrors.New(code.AuthBlank, code.AuthBlankMsg)
	}
	couponUsages, err := l.svcCtx.CouponRPC.ListCouponUsages(l.ctx, &couponsclient.ListCouponUsagesReq{
		UserId: userID,
		Pagination: &couponsclient.PaginationReq{
			Page: req.Page,
			Size: req.PageSize,
		},
	})
	if err != nil {
		l.Logger.Errorw("call rpc ListCouponUsages failed", logx.Field("err", err))
		return nil, xerrors.New(code.ServerError, code.ServerErrorMsg)
	}
	if couponUsages.StatusCode != code.Success {
		return nil, xerrors.New(int(couponUsages.StatusCode), couponUsages.StatusMsg)
	}
	resp = &types.CouponUsageListResp{
		CouponUsageList: convertCouponUsageList2Resp(couponUsages.Usages),
	}
	return
}
