package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/svc"
	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/core/logx"
	xerrors "github.com/zeromicro/x/errors"
)

// FlashBuyLogic is the business logic for FlashBuyLogic operations.
type FlashBuyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFlashBuyLogic creates a new FlashBuyLogic instance.
func NewFlashBuyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FlashBuyLogic {
	return &FlashBuyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// does something.
func (l *FlashBuyLogic) FlashBuy(req *types.FlashBuyReq) (resp *types.FlashBuyResp, err error) {
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, xerrors.New(code.AuthBlank, code.AuthBlankMsg)
	}

	// 0. 获取用户地址
	addressListResp, err := l.svcCtx.UsersRPC.ListAddresses(l.ctx, &users.AllAddressLitstRequest{
		UserId: userID,
	})
	if err != nil {
		l.Logger.Errorw("get user addresses failed", logx.Field("err", err))
		return nil, xerrors.New(code.ServerError, "获取收货地址失败，请稍后重试")
	}
	if addressListResp == nil || len(addressListResp.Data) == 0 {
		return nil, xerrors.New(code.ServerError, "请先添加收货地址")
	}

	var addressID uint32
	// 优先使用默认地址
	for _, addr := range addressListResp.Data {
		if addr.IsDefault {
			addressID = uint32(addr.AddressId)
			break
		}
	}
	// 如果没有默认地址，使用第一个
	if addressID == 0 {
		addressID = uint32(addressListResp.Data[0].AddressId)
	}

	// 1. 准备结算（模拟购物车结算过程）
	checkoutReq := &checkout.CheckoutReq{
		UserId:   userID,
		CouponId: "",
		OrderItems: []*checkout.CheckoutReq_OrderItem{
			{
				ProductId: int32(req.ProductID),
				Quantity:  int32(req.Quantity),
			},
		},
		AddressId: uint64(addressID),
	}

	checkoutResp, err := l.svcCtx.CheckoutRPC.PrepareCheckout(l.ctx, checkoutReq)
	if err != nil {
		l.Logger.Errorw("prepare checkout failed", logx.Field("err", err))
		return nil, xerrors.New(code.ServerError, code.ServerErrorMsg)
	}
	if checkoutResp.StatusCode != code.Success {
		return nil, xerrors.New(int(checkoutResp.StatusCode), checkoutResp.StatusMsg)
	}

	// 2. 创建订单
	createOrderReq := &order.CreateOrderRequest{
		PreOrderId:    checkoutResp.PreOrderId,
		UserId:        userID,
		PaymentMethod: order.PaymentMethod_ALIPAY,
		AddressId:     uint64(addressID),
		CouponId:      "",
	}

	createOrderResp, err := l.svcCtx.OrderRPC.CreateOrder(l.ctx, createOrderReq)
	if err != nil {
		l.Logger.Errorw("create order failed", logx.Field("err", err))
		return nil, xerrors.New(code.ServerError, code.ServerErrorMsg)
	}
	if createOrderResp.StatusCode != code.Success {
		return nil, xerrors.New(int(createOrderResp.StatusCode), createOrderResp.StatusMsg)
	}

	// 检查是否返回了订单信息
	if createOrderResp.Order == nil {
		l.Logger.Errorw("create order returned nil order")
		return nil, xerrors.New(code.ServerError, "创建订单失败")
	}

	return &types.FlashBuyResp{
		OrderID: createOrderResp.Order.OrderId,
		OrderNo: createOrderResp.Order.OrderId,
		Total:   createOrderResp.Order.PayableAmount,
	}, nil
}
