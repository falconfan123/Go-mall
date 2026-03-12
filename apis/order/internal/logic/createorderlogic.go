package logic

import (
	"context"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/falconfan123/Go-mall/apis/order/internal/svc"
	"github.com/falconfan123/Go-mall/apis/order/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/services/coupons/pb"
	"github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	xerrors "github.com/zeromicro/x/errors"
	"strings"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.CreateOrderResp, err error) {
	l.Logger.Infof("CreateOrder called. Req: %+v", req)
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, xerrors.New(code.AuthBlank, code.AuthBlankMsg)
	}

	// DTM在docker容器内，无法访问localhost:8500，需要替换为consul:8500
	// 并且由于RPC服务在宿主机，DTM需要通过host.docker.internal访问
	// 这里直接使用直连方式，绕过Consul解析
	orderTarget := "direct://host.docker.internal:10004"

	checkoutTarget := "direct://host.docker.internal:10005"

	couponTarget := "direct://host.docker.internal:10009"

	// --------------- saga ---------------
	// 去掉direct://前缀
	dtmTarget := strings.TrimPrefix(l.svcCtx.Config.DtmRpc.Target, "direct://")
	sagaGrpc := dtmgrpc.NewSagaGrpc(dtmTarget, uuid.New().String())
	orderID := uuid.New().String()
	if req.CouponID != "" {
		// 锁定优惠券
		sagaGrpc.Add(couponTarget+coupons.Coupons_LockCoupon_FullMethodName,
			couponTarget+coupons.Coupons_ReleaseCoupon_FullMethodName, &coupons.LockCouponReq{
				UserId:       int32(userID),
				UserCouponId: req.CouponID,
				PreOrderId:   req.PreOrderID,
			})
	}
	// 锁定结算，进入结算确认状态
	sagaGrpc.Add(checkoutTarget+checkout.CheckoutService_UpdateStatus2Order_FullMethodName,
		checkoutTarget+checkout.CheckoutService_UpdateStatus2OrderRollback_FullMethodName, &checkout.UpdateStatusReq{
			UserId:     int32(userID),
			PreOrderId: req.PreOrderID,
		}).
		// 创建订单
		Add(orderTarget+order.OrderService_CreateOrder_FullMethodName,
			orderTarget+order.OrderService_CreateOrderRollback_FullMethodName, &order.CreateOrderRequest{
				UserId:        userID,
				PreOrderId:    req.PreOrderID,
				PaymentMethod: order.PaymentMethod_ALIPAY,
				AddressId:     uint64(req.AddressID),
				CouponId:      req.CouponID,
				OrderId:       orderID,
			})
	sagaGrpc.WithGlobalTransRequestTimeout(5000)
	sagaGrpc.WaitResult = true // 等待结果
	l.Logger.Infof("Submitting Saga transaction with OrderID: %s", orderID)
	if err := sagaGrpc.Submit(); err != nil {
		l.Logger.Errorw("call rpc Submit failed", logx.Field("err", err))
		return nil, xerrors.New(code.CreateOrderFailed, code.CreateOrderFailedMsg)
	}
	l.Logger.Infof("Saga transaction submitted successfully, OrderID: %s", orderID)
	return &types.CreateOrderResp{
		OrderID: orderID,
	}, nil
}
