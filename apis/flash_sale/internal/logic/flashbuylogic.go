package logic

import (
	"context"
	"jijizhazha1024/go-mall/apis/flash_sale/internal/svc"
	"jijizhazha1024/go-mall/apis/flash_sale/internal/types"
	"jijizhazha1024/go-mall/common/consts/biz"
	"jijizhazha1024/go-mall/common/consts/code"
	"jijizhazha1024/go-mall/services/checkout/checkout"
	"jijizhazha1024/go-mall/services/order/order"
	"jijizhazha1024/go-mall/services/users/usersclient"

	"github.com/zeromicro/go-zero/core/logx"
	xerrors "github.com/zeromicro/x/errors"
)

type FlashBuyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFlashBuyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FlashBuyLogic {
	return &FlashBuyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FlashBuyLogic) FlashBuy(req *types.FlashBuyReq) (resp *types.FlashBuyResp, err error) {
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, xerrors.New(code.AuthBlank, code.AuthBlankMsg)
	}

	// 0. 获取用户地址
	addressListResp, err := l.svcCtx.UsersRpc.ListAddresses(l.ctx, &usersclient.AllAddressLitstRequest{
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

	checkoutResp, err := l.svcCtx.CheckoutRpc.PrepareCheckout(l.ctx, checkoutReq)
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

	createOrderResp, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, createOrderReq)
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

func newError(msg string) error {
	return &customError{msg}
}

type customError struct {
	message string
}

func (e *customError) Error() string {
	return e.message
}
