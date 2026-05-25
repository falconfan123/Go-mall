package checkout

import (
	"context"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"testing"
)

func TestGetCheckoutDetail(t *testing.T) {
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

	detail, err := clients.Checkout.GetCheckoutDetail(context.TODO(), &checkout.CheckoutDetailReq{
		PreOrderId: prepareResp.GetPreOrderId(),
		UserId:     int32(user.UserID),
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(detail)
}
func TestGetCheckoutList(t *testing.T) {
	clients := harness.NewClients(t)
	harness.WaitForServices(t, clients)

	user := seed.CreateUser(t, clients.Users)
	product := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 8)
	if _, err := clients.Checkout.PrepareCheckout(context.TODO(), seed.MakeCheckoutRequest(
		user.UserID,
		user.AddressID,
		&checkout.CheckoutReq_OrderItem{ProductId: int32(product.ProductID), Quantity: 1},
	)); err != nil {
		t.Fatal(err)
	}

	list, err := clients.Checkout.GetCheckoutList(context.TODO(), &checkout.CheckoutListReq{
		PageSize: 5,
		Page:     1,
		UserId:   user.UserID,
	})
	if err != nil {
		t.Error(err)
	}
	for _, v := range list.Data {
		t.Log(v)
	}
}
