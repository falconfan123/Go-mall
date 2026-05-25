package order

import (
	"context"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"testing"
	"time"
)

func TestUpdateOrder(t *testing.T) {
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
	createResp, err := clients.Order.CreateOrder(context.Background(), seed.MakeOrderRequest(
		checkoutResp.GetPreOrderId(),
		user.UserID,
		user.AddressID,
		order.PaymentMethod_ALIPAY,
		testenv.UniqueName("order"),
	))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("订单状态更新为支付中与补偿", func(t *testing.T) {
		res, err := clients.Order.UpdateOrder2PaymentStatus(context.Background(), &order.UpdateOrder2PaymentRequest{
			OrderId: createResp.GetOrder().GetOrderId(),
			UserId:  int32(user.UserID),
		})
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(res)
		rollback, err := clients.Order.UpdateOrder2PaymentStatusRollback(context.Background(), &order.UpdateOrder2PaymentRequest{
			OrderId: createResp.GetOrder().GetOrderId(),
			UserId:  int32(user.UserID),
		})
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(rollback)

	})
	t.Run("订单状态更新为支付成功与补偿", func(t *testing.T) {
		res, err := clients.Order.UpdateOrder2PaymentSuccess(context.Background(), &order.UpdateOrder2PaymentSuccessRequest{
			OrderId: createResp.GetOrder().GetOrderId(),
			UserId:  int32(user.UserID),
			PaymentResult: &order.PaymentResult{
				TransactionId: "xxxx",
				PaidAmount:    100,
				PaidAt:        time.Now().Unix(),
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(res)
		rollback, err := clients.Order.UpdateOrder2PaymentSuccessRollback(context.Background(), &order.UpdateOrder2PaymentSuccessRequest{
			OrderId: createResp.GetOrder().GetOrderId(),
			UserId:  int32(user.UserID),
			PaymentResult: &order.PaymentResult{
				TransactionId: "xxxx",
				PaidAmount:    100,
				PaidAt:        time.Now().Unix(),
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(rollback)
	})
}

func TestDtmSaga(t *testing.T) {
	t.Skip("dtm saga test is optional in the default integration suite")
}
