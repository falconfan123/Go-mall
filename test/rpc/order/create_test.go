package order

import (
	"context"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	order "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"testing"
)

func TestCreateOrder(t *testing.T) {
	clients := harness.NewClients(t)
	harness.WaitForServices(t, clients)

	user := seed.CreateUser(t, clients.Users)
	product := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 8)
	checkoutResp, err := clients.Checkout.PrepareCheckout(context.Background(), seed.MakeCheckoutRequest(
		user.UserID,
		user.AddressID,
		&checkout.CheckoutReq_OrderItem{ProductId: int32(product.ProductID), Quantity: 1},
	))
	if err != nil {
		t.Fatal(err)
	}

	createOrder, err := clients.Order.CreateOrder(context.TODO(), &order.CreateOrderRequest{
		PreOrderId:    checkoutResp.GetPreOrderId(),
		UserId:        user.UserID,
		AddressId:     user.AddressID,
		PaymentMethod: order.PaymentMethod_ALIPAY,
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(createOrder)
}
