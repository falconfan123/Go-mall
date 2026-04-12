package order

import (
	"context"
	"github.com/falconfan123/Go-mall/services/order/pb"
	"testing"
)

func TestCreateOrder(t *testing.T) {
	createOrder, err := orderClient.CreateOrder(context.TODO(), &order.CreateOrderRequest{
		PreOrderId:    "019555d7-8dca-7f17-b945-cee24c0efb7b",
		UserId:        1,
		AddressId:     1,
		PaymentMethod: order.PaymentMethod_ALIPAY,
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(createOrder)
}
