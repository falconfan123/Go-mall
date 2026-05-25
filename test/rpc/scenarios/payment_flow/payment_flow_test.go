package payment_flow

import (
	"context"
	"testing"
	"time"

	checkoutpb "github.com/falconfan123/Go-mall/services/checkout/pb"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	paymentpb "github.com/falconfan123/Go-mall/services/payment/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"github.com/stretchr/testify/assert"
)

func TestPaymentHappyPath(t *testing.T) {
	clients := harness.NewClients(t)
	harness.WaitForServices(t, clients)

	user := seed.CreateUser(t, clients.Users)
	product := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 12)

	ctx, cancel := context.WithTimeout(context.Background(), testenv.Timeout())
	defer cancel()

	checkoutResp, err := clients.Checkout.PrepareCheckout(ctx, seed.MakeCheckoutRequest(
		user.UserID,
		user.AddressID,
		&checkoutpb.CheckoutReq_OrderItem{ProductId: int32(product.ProductID), Quantity: 1},
	))
	assert.NoError(t, err)

	orderID := testenv.UniqueName("order")
	orderResp, err := clients.Order.CreateOrder(ctx, seed.MakeOrderRequest(
		checkoutResp.GetPreOrderId(),
		user.UserID,
		user.AddressID,
		orderpb.PaymentMethod_ALIPAY,
		orderID,
	))
	assert.NoError(t, err)
	assert.NotNil(t, orderResp.GetOrder())

	paymentResp, err := clients.Payment.CreatePayment(ctx, &paymentpb.PaymentReq{
		UserId:        user.UserID,
		OrderId:       orderID,
		PaymentMethod: paymentpb.PaymentMethod_ALIPAY,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, int(paymentResp.GetStatusCode()))
	assert.NotNil(t, paymentResp.GetPayment())
	assert.Equal(t, orderID, paymentResp.GetPayment().GetOrderId())

	snapshot, err := clients.Order.GetOrder2Payment(ctx, &orderpb.GetOrderRequest{
		OrderId: orderID,
		UserId:  user.UserID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, snapshot.GetOrder())

	if snapshot.GetOrder().GetOrderStatus() != orderpb.OrderStatus_ORDER_STATUS_PENDING_PAYMENT {
		_, err = clients.Order.UpdateOrder2PaymentStatus(ctx, &orderpb.UpdateOrder2PaymentRequest{
			OrderId: orderID,
			UserId:  int32(user.UserID),
		})
		assert.NoError(t, err)
	}

	_, err = clients.Order.UpdateOrder2PaymentSuccess(ctx, &orderpb.UpdateOrder2PaymentSuccessRequest{
		OrderId: orderID,
		UserId:  int32(user.UserID),
		PaymentResult: &orderpb.PaymentResult{
			TransactionId: testenv.UniqueName("txn"),
			PaidAmount:    snapshot.GetOrder().GetPayableAmount(),
			PaidAt:        time.Now().Unix(),
		},
	})
	assert.NoError(t, err)

	finalOrder, err := clients.Order.GetOrder2Payment(ctx, &orderpb.GetOrderRequest{
		OrderId: orderID,
		UserId:  user.UserID,
	})
	assert.NoError(t, err)
	assert.Equal(t, orderpb.OrderStatus_ORDER_STATUS_PAID, finalOrder.GetOrder().GetOrderStatus())
}
