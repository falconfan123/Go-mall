package checkout

import (
	"context"
	"github.com/falconfan123/Go-mall/services/checkout/pb"
	"testing"
)

func TestRelease(t *testing.T) {
	resp, err := checkoutClient.ReleaseCheckout(context.TODO(), &checkout.ReleaseReq{
		PreOrderId: "019554a5-838c-7414-868d-aba6c1d7c6cd", UserId: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
