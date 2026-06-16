package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	ordertypes "github.com/falconfan123/Go-mall/common/types/order"
	ordermodel "github.com/falconfan123/Go-mall/dal/model/order"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/services/order/internal/svc"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	userspb "github.com/falconfan123/Go-mall/services/users/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *orderpb.CreateOrderRequest) (*orderpb.OrderDetailResponse, error) {
	res := &orderpb.OrderDetailResponse{}
	if in.PreOrderId == "" || in.UserId == 0 || in.AddressId == 0 {
		res.StatusCode = code.OrderParameterInvalid
		res.StatusMsg = code.OrderParameterInvalidMsg
		return res, nil
	}

	orderID, err := l.svcCtx.OrderModel.GetOrderIDByPreID(l.ctx, in.PreOrderId, int32(in.UserId))
	if err != nil {
		l.Logger.Errorw("get existing order id failed",
			logx.Field("err", err),
			logx.Field("pre_order_id", in.PreOrderId),
			logx.Field("user_id", in.UserId))
		return nil, status.Error(codes.Internal, code.CreateOrderFailedMsg)
	}
	if orderID != "" {
		return l.loadOrderDetail(orderID, in.UserId)
	}

	checkoutDetail, err := l.svcCtx.CheckoutRpc.GetCheckoutDetail(l.ctx, &checkout.CheckoutDetailReq{
		PreOrderId: in.PreOrderId,
		UserId:     int32(in.UserId),
	})
	if err != nil {
		l.Logger.Errorw("get checkout detail failed",
			logx.Field("err", err),
			logx.Field("pre_order_id", in.PreOrderId),
			logx.Field("user_id", in.UserId))
		return nil, err
	}
	if checkoutDetail.StatusCode != code.Success || checkoutDetail.Data == nil {
		res.StatusCode = checkoutDetail.StatusCode
		if res.StatusCode == code.Success {
			res.StatusCode = code.CreateOrderFailed
		}
		res.StatusMsg = checkoutDetail.StatusMsg
		if res.StatusMsg == "" {
			res.StatusMsg = code.CreateOrderFailedMsg
		}
		return res, nil
	}

	addressResp, err := l.svcCtx.UserRpc.GetAddress(l.ctx, &userspb.GetAddressRequest{
		UserId:    in.UserId,
		AddressId: in.AddressId,
	})
	if err != nil {
		l.Logger.Errorw("get address failed",
			logx.Field("err", err),
			logx.Field("address_id", in.AddressId),
			logx.Field("user_id", in.UserId))
		return nil, err
	}
	if addressResp.GetStatusCode() != code.Success || addressResp.GetData() == nil {
		res.StatusCode = int32(addressResp.GetStatusCode())
		if res.StatusCode == code.Success {
			res.StatusCode = code.UserOrderAddressNotExist
		}
		res.StatusMsg = addressResp.GetStatusMsg()
		if res.StatusMsg == "" {
			res.StatusMsg = code.UserOrderAddressNotExistMsg
		}
		return res, nil
	}

	newOrderID := in.OrderId
	if newOrderID == "" {
		newOrderID = in.PreOrderId
	}

	orderModel := &ordermodel.Orders{
		OrderId:        newOrderID,
		PreOrderId:     in.PreOrderId,
		UserId:         uint64(in.UserId),
		CouponId:       in.CouponId,
		PaymentMethod:  sql.NullInt64{Int64: int64(in.PaymentMethod), Valid: in.PaymentMethod != orderpb.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED},
		OriginalAmount: checkoutDetail.Data.OriginalAmount,
		DiscountAmount: checkoutDetail.Data.OriginalAmount - checkoutDetail.Data.FinalAmount,
		PayableAmount:  checkoutDetail.Data.FinalAmount,
		OrderStatus:    int64(ordertypes.OrderStatusCreated),
		PaymentStatus:  int64(ordertypes.PaymentStatusNotPaid),
		ExpireTime:     checkoutDetail.Data.ExpireTime,
	}

	orderItems := make([]*ordermodel.OrderItems, len(checkoutDetail.Data.Items))
	respItems := make([]*orderpb.OrderItem, len(checkoutDetail.Data.Items))
	for i, item := range checkoutDetail.Data.Items {
		orderItems[i] = &ordermodel.OrderItems{
			OrderId:     newOrderID,
			ProductId:   uint64(item.ProductId),
			Quantity:    uint64(item.Quantity),
			Price:       item.Price,
			ProductName: item.ProductName,
			ProductDesc: item.ProductDesc,
		}
		respItems[i] = &orderpb.OrderItem{
			ProductId:   uint64(item.ProductId),
			Quantity:    uint64(item.Quantity),
			ProductName: item.ProductName,
			ProductDesc: item.ProductDesc,
			UnitPrice:   item.Price,
		}
	}

	orderAddress := &ordermodel.OrderAddresses{
		OrderId:         newOrderID,
		AddressId:       addressResp.Data.AddressId,
		RecipientName:   addressResp.Data.RecipientName,
		PhoneNumber:     sql.NullString{String: addressResp.Data.PhoneNumber, Valid: addressResp.Data.PhoneNumber != ""},
		Province:        sql.NullString{String: addressResp.Data.Province, Valid: addressResp.Data.Province != ""},
		City:            addressResp.Data.City,
		DetailedAddress: addressResp.Data.DetailedAddress,
	}

	if err := l.svcCtx.Model.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := l.svcCtx.OrderModel.WithSession(session).Insert(ctx, orderModel); err != nil {
			return err
		}
		if len(orderItems) > 0 {
			if err := l.svcCtx.OrderItemModel.WithSession(session).BulkInsert(session, orderItems); err != nil {
				return err
			}
		}
		if _, err := l.svcCtx.OrderAddress.WithSession(session).Insert(ctx, orderAddress); err != nil {
			return err
		}
		updateResp, err := l.svcCtx.CheckoutRpc.UpdateStatus2Order(ctx, &checkout.UpdateStatusReq{
			PreOrderId: in.PreOrderId,
			UserId:     int32(in.UserId),
		})
		if err != nil {
			return err
		}
		if updateResp.GetStatusCode() != code.Success {
			return status.Error(codes.Aborted, updateResp.GetStatusMsg())
		}
		return nil
	}); err != nil {
		l.Logger.Errorw("create order transaction failed",
			logx.Field("err", err),
			logx.Field("order_id", orderID),
			logx.Field("pre_order_id", in.PreOrderId),
			logx.Field("user_id", in.UserId))
		return nil, status.Error(codes.Internal, code.CreateOrderFailedMsg)
	}

	res.StatusCode = code.Success
	res.StatusMsg = code.SuccessMsg
	res.Order = &orderpb.Order{
		OrderId:        newOrderID,
		PreOrderId:     in.PreOrderId,
		UserId:         in.UserId,
		PaymentMethod:  in.PaymentMethod,
		OriginalAmount: checkoutDetail.Data.OriginalAmount,
		DiscountAmount: checkoutDetail.Data.OriginalAmount - checkoutDetail.Data.FinalAmount,
		PayableAmount:  checkoutDetail.Data.FinalAmount,
		OrderStatus:    orderpb.OrderStatus_ORDER_STATUS_CREATED,
		PaymentStatus:  orderpb.PaymentStatus_PAYMENT_STATUS_NOT_PAID,
		ExpireTime:     time.Unix(checkoutDetail.Data.ExpireTime, 0).Format(time.DateTime),
		Items:          respItems,
	}
	res.Items = respItems
	res.Address = &orderpb.OrderAddress{
		AddressId:       addressResp.Data.AddressId,
		RecipientName:   addressResp.Data.RecipientName,
		PhoneNumber:     addressResp.Data.PhoneNumber,
		Province:        addressResp.Data.Province,
		City:            addressResp.Data.City,
		DetailedAddress: addressResp.Data.DetailedAddress,
		CreatedAt:       addressResp.Data.CreatedAt,
		UpdatedAt:       addressResp.Data.UpdatedAt,
		OrderId:         newOrderID,
	}
	return res, nil
}

func (l *CreateOrderLogic) loadOrderDetail(orderID string, userID uint32) (*orderpb.OrderDetailResponse, error) {
	getLogic := NewGetOrderLogic(l.ctx, l.svcCtx)
	return getLogic.GetOrder(&orderpb.GetOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	})
}
