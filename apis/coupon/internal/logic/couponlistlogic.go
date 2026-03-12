package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/svc"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/coupons/pb"
	"github.com/zeromicro/x/errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type CouponListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCouponListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CouponListLogic {
	return &CouponListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CouponListLogic) CouponList(req *types.CouponListReq) (resp *types.CouponListResp, err error) {

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > biz.MaxPageSize {
		req.PageSize = biz.MaxPageSize
	}
	res, err := l.svcCtx.CouponRpc.ListCoupons(l.ctx, &couponsclient.ListCouponsReq{
		Pagination: &couponsclient.PaginationReq{
			Page: req.Page,
			Size: req.PageSize,
		},
		Type: int32(req.Type),
	})
	if err != nil {
		if res != nil && res.StatusCode != code.Success {
			// 处理用户级别info 错误
			return nil, errors.New(int(res.StatusCode), res.StatusMsg)
		}
		l.Logger.Errorw("call rpc ListCoupons failed", logx.Field("err", err))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	}
	if res.StatusCode != code.Success {
		return nil, errors.New(int(res.StatusCode), res.StatusMsg)
	}
	resp = &types.CouponListResp{
		CouponList: make([]types.CouponItemResp, len(res.Coupons)),
	}
	for i, item := range res.Coupons {
		resp.CouponList[i] = *convertCoupon2Resp(item)
	}
	return
}
