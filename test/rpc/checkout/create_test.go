package checkout

import (
	"context"
	"github.com/falconfan123/Go-mall/services/checkout/pb"
	"testing"
)

func TestPrePareCheckout(t *testing.T) {
	resp, err := checkoutClient.PrepareCheckout(context.TODO(), &checkout.CheckoutReq{
		UserId: 1,
		OrderItems: []*checkout.CheckoutReq_OrderItem{
			{
				ProductId: 11,
				Quantity:  1,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}
