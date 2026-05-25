package checkout_flow

import (
	"context"
	"testing"

	checkoutpb "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"github.com/stretchr/testify/assert"
)

func TestCheckoutHappyPath(t *testing.T) {
	clients := harness.NewClients(t)
	harness.WaitForServices(t, clients)

	user := seed.CreateUser(t, clients.Users)
	product := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 8)

	ctx, cancel := context.WithTimeout(context.Background(), testenv.Timeout())
	defer cancel()

	resp, err := clients.Checkout.PrepareCheckout(ctx, seed.MakeCheckoutRequest(
		user.UserID,
		user.AddressID,
		&checkoutpb.CheckoutReq_OrderItem{ProductId: int32(product.ProductID), Quantity: 2},
	))
	assert.NoError(t, err)
	assert.Equal(t, 0, int(resp.GetStatusCode()))
	assert.NotEmpty(t, resp.GetPreOrderId())
	assert.Greater(t, len(resp.GetPayMethod()), 0)

	detail, err := clients.Checkout.GetCheckoutDetail(ctx, &checkoutpb.CheckoutDetailReq{
		PreOrderId: resp.GetPreOrderId(),
		UserId:     int32(user.UserID),
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, int(detail.GetStatusCode()))
	assert.NotNil(t, detail.GetData())
	assert.Equal(t, resp.GetPreOrderId(), detail.GetData().GetPreOrderId())
}
