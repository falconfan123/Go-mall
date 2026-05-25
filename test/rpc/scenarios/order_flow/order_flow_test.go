package order_flow

import (
	"context"
	"testing"

	checkoutpb "github.com/falconfan123/Go-mall/services/checkout/pb"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"github.com/stretchr/testify/assert"
)

func TestOrderHappyPath(t *testing.T) {
	clients := harness.NewClients(t)
	harness.WaitForServices(t, clients)

	user := seed.CreateUser(t, clients.Users)
	product := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 10)

	ctx, cancel := context.WithTimeout(context.Background(), testenv.Timeout())
	defer cancel()

	checkoutResp, err := clients.Checkout.PrepareCheckout(ctx, seed.MakeCheckoutRequest(
		user.UserID,
		user.AddressID,
		&checkoutpb.CheckoutReq_OrderItem{ProductId: int32(product.ProductID), Quantity: 1},
	))
	assert.NoError(t, err)
	assert.NotEmpty(t, checkoutResp.GetPreOrderId())

	orderID := testenv.UniqueName("order")
	orderResp, err := clients.Order.CreateOrder(ctx, seed.MakeOrderRequest(
		checkoutResp.GetPreOrderId(),
		user.UserID,
		user.AddressID,
		orderpb.PaymentMethod_ALIPAY,
		orderID,
	))
	assert.NoError(t, err)
	assert.Equal(t, 0, int(orderResp.GetStatusCode()))
	assert.NotNil(t, orderResp.GetOrder())
	assert.Equal(t, orderID, orderResp.GetOrder().GetOrderId())
	assert.Equal(t, checkoutResp.GetPreOrderId(), orderResp.GetOrder().GetPreOrderId())
	assert.Equal(t, user.UserID, orderResp.GetOrder().GetUserId())
	assert.Equal(t, orderpb.PaymentMethod_ALIPAY, orderResp.GetOrder().GetPaymentMethod())
	assert.Greater(t, orderResp.GetOrder().GetPayableAmount(), int64(0))
	assert.Len(t, orderResp.GetItems(), 1)
}
