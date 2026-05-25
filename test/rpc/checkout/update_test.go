package checkout

import (
	"context"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"testing"
)

func TestUpdateStatus(t *testing.T) {
	clients := harness.NewClients(t)
	harness.WaitForServices(t, clients)

	user := seed.CreateUser(t, clients.Users)
	product := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 8)
	prepareResp, err := clients.Checkout.PrepareCheckout(context.TODO(), seed.MakeCheckoutRequest(
		user.UserID,
		user.AddressID,
		&checkout.CheckoutReq_OrderItem{ProductId: int32(product.ProductID), Quantity: 1},
	))
	if err != nil {
		t.Fatal(err)
	}

	order, err := clients.Checkout.UpdateStatus2Order(context.TODO(), &checkout.UpdateStatusReq{
		PreOrderId: prepareResp.GetPreOrderId(),
		UserId:     int32(user.UserID),
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(order)
}
