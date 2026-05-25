package checkout

import (
	"context"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"testing"
)

func TestPrePareCheckout(t *testing.T) {
	clients := harness.NewClients(t)
	harness.WaitForServices(t, clients)

	user := seed.CreateUser(t, clients.Users)
	product := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 8)

	resp, err := clients.Checkout.PrepareCheckout(context.TODO(), seed.MakeCheckoutRequest(
		user.UserID,
		user.AddressID,
		&checkout.CheckoutReq_OrderItem{
			ProductId: int32(product.ProductID),
			Quantity:  1,
		},
	))
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}
