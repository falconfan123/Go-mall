package logic

import (
	"context"
	"errors"

	"github.com/falconfan123/Go-mall/common/consts/code"
	paymentmodel "github.com/falconfan123/Go-mall/dal/model/payment"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	payment "github.com/falconfan123/Go-mall/services/payment/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentStatusLogic {
	return &GetPaymentStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPaymentStatusLogic) GetPaymentStatus(in *payment.PaymentStatusReq) (*payment.PaymentStatusResp, error) {
	res := &payment.PaymentStatusResp{}
	if in.PaymentId == "" {
		res.StatusCode = code.OrderParameterInvalid
		res.StatusMsg = code.OrderParameterInvalidMsg
		return res, nil
	}

	record, err := l.svcCtx.PaymentModel.FindOne(l.ctx, in.PaymentId)
	if err != nil {
		if errors.Is(err, paymentmodel.ErrNotFound) {
			res.StatusCode = code.PaymentNotExist
			res.StatusMsg = code.PaymentNotExistMsg
			return res, nil
		}
		l.Logger.Errorw("query payment failed", logx.Field("err", err), logx.Field("payment_id", in.PaymentId))
		return nil, err
	}

	if in.OrderId != "" && record.OrderId.String != "" && in.OrderId != record.OrderId.String {
		res.StatusCode = code.OrderParameterInvalid
		res.StatusMsg = "支付单与订单号不匹配"
		return res, nil
	}

	res.Payment = ConvertModelToPaymentItem(record)

	orderID := record.OrderId.String
	if orderID == "" && in.OrderId != "" {
		orderID = in.OrderId
	}
	if orderID == "" || record.UserId == 0 {
		return res, nil
	}

	orderResp, err := l.svcCtx.OrderRpc.GetOrder2Payment(l.ctx, &orderpb.GetOrderRequest{
		OrderId: orderID,
		UserId:  uint32(record.UserId),
	})
	if err != nil {
		l.Logger.Errorw("query order payment status failed",
			logx.Field("err", err),
			logx.Field("payment_id", in.PaymentId),
			logx.Field("order_id", orderID))
		return nil, err
	}
	if orderResp.StatusCode != code.Success || orderResp.Order == nil {
		res.StatusCode = orderResp.StatusCode
		res.StatusMsg = orderResp.StatusMsg
		return res, nil
	}

	res.OrderStatus = int32(orderResp.Order.OrderStatus)
	res.OrderPaymentStatus = int32(orderResp.Order.PaymentStatus)
	return res, nil
}
