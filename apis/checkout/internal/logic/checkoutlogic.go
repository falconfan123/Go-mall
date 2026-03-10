package logic

import (
	"context"
	xerrors "github.com/zeromicro/x/errors"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/checkout/checkout"

	"github.com/falconfan123/Go-mall/apis/checkout/internal/svc"
	"github.com/falconfan123/Go-mall/apis/checkout/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckoutLogic {
	return &CheckoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckoutLogic) Checkout(req *types.CheckoutReq) (resp *types.CheckoutResp, err error) {
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, xerrors.New(code.AuthBlank, code.AuthBlankMsg)
	}

	// 解析地址ID
	var addressID uint64
	if req.AddressID != 0 {
		addressID = uint64(req.AddressID)
	}

	res, err := l.svcCtx.CheckoutRpc.PrepareCheckout(l.ctx, &checkout.CheckoutReq{
		UserId:     userID,
		CouponId:   req.CouponID,
		OrderItems: convertCheckoutItem2Req(req.OrderItems),
		AddressId:  addressID,
	})
	if err != nil {
		l.Logger.Errorw("call rpc GetOrder failed", logx.Field("err", err))
		return nil, xerrors.New(code.ServerError, code.ServerErrorMsg)
	}
	if res.StatusCode != code.Success {
		return nil, xerrors.New(int(res.StatusCode), res.StatusMsg)
	}
	resp = &types.CheckoutResp{
		ExpireTime: res.ExpireTime,
		PayMethod:  res.PayMethod,
		PreOrderID: res.PreOrderId,
	}
	return
}

func convertCheckoutItem2Req(items []types.CheckoutItemReq) []*checkout.CheckoutReq_OrderItem {
	orderItems := make([]*checkout.CheckoutReq_OrderItem, len(items))
	for i, item := range items {
		orderItems[i] = &checkout.CheckoutReq_OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return orderItems
}
