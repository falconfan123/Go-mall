package logic

import (
	order2 "github.com/falconfan123/Go-mall/dal/model/order"
	checkoutpb "github.com/falconfan123/Go-mall/services/checkout/pb"
	couponspb "github.com/falconfan123/Go-mall/services/coupons/pb"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	"time"
)

func convertToCouponItems(items []*checkoutpb.CheckoutItem) []*couponspb.Items {
	couponItems := make([]*couponspb.Items, len(items))
	for i, item := range items {
		couponItems[i] = &couponspb.Items{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		}
	}
	return couponItems
}
func convertToOrderItems(orderID string, items []*checkoutpb.CheckoutItem) []*order2.OrderItems {
	orderItems := make([]*order2.OrderItems, len(items))
	for i, item := range items {
		orderItems[i] = &order2.OrderItems{
			OrderId:   orderID,
			ProductId: uint64(item.ProductId),
			Quantity:  uint64(item.Quantity),
			Price:     item.Price,
		}
	}
	return orderItems
}

// --------------- resp ---------------
func convertToOrderResp(orderModelRes *order2.Orders) *orderpb.Order {
	resp := &orderpb.Order{
		OrderId:        orderModelRes.OrderId,
		OrderStatus:    orderpb.OrderStatus(orderModelRes.OrderStatus),
		PaymentStatus:  orderpb.PaymentStatus(orderModelRes.PaymentStatus),
		PaymentMethod:  orderpb.PaymentMethod(orderModelRes.PaymentMethod.Int64),
		OriginalAmount: orderModelRes.OriginalAmount,
		PayableAmount:  orderModelRes.PayableAmount,
		PaidAmount:     orderModelRes.PaidAmount.Int64,
		PaidAt:         orderModelRes.PaidAt.Int64,
		DiscountAmount: orderModelRes.DiscountAmount,
		ExpireTime:     time.Unix(orderModelRes.ExpireTime, 0).Format(time.DateTime),
		CreatedAt:      orderModelRes.CreatedAt.Format(time.DateTime),
		UpdatedAt:      orderModelRes.UpdatedAt.Format(time.DateTime),
		PreOrderId:     orderModelRes.PreOrderId,
		Reason:         orderModelRes.Reason.String,
		TransactionId:  orderModelRes.TransactionId.String,
		UserId:         uint32(orderModelRes.UserId),
		Items:          []*orderpb.OrderItem{}, // 初始化订单项切片
	}

	return resp
}

func convertToOrderItemResp(orderItems []*order2.OrderItems) []*orderpb.OrderItem {
	resp := make([]*orderpb.OrderItem, len(orderItems))
	for i, item := range orderItems {

		resp[i] = &orderpb.OrderItem{
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			UnitPrice:   item.Price,
			Quantity:    item.Quantity,
			ProductDesc: item.ProductDesc,
		}
	}
	return resp
}
func convertToOrderAddressResp(address *order2.OrderAddresses) *orderpb.OrderAddress {
	return &orderpb.OrderAddress{
		AddressId:       address.AddressId,
		RecipientName:   address.RecipientName,
		PhoneNumber:     address.PhoneNumber.String,
		Province:        address.Province.String,
		City:            address.City,
		DetailedAddress: address.DetailedAddress,
		OrderId:         address.OrderId,
		CreatedAt:       address.CreatedAt.Format(time.DateTime),
		UpdatedAt:       address.UpdatedAt.Format(time.DateTime),
	}
}
