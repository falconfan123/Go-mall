package order

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/code"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"testing"
)

func TestListOrders(t *testing.T) {
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
	if _, err := clients.Order.CreateOrder(context.Background(), seed.MakeOrderRequest(
		checkoutResp.GetPreOrderId(),
		user.UserID,
		user.AddressID,
		order.PaymentMethod_ALIPAY,
		testenv.UniqueName("order"),
	)); err != nil {
		t.Fatal(err)
	}

	t.Run("获取订单列表", func(t *testing.T) {
		listOrders, err := clients.Order.ListOrders(context.Background(), &order.ListOrdersRequest{
			Pagination: &order.ListOrdersRequest_Pagination{
				Page:     1,
				PageSize: 10,
			},
			UserId: user.UserID,
		})
		if err != nil {
			t.Error(err)
		}
		if listOrders.StatusCode != code.Success {
			t.Log(listOrders.StatusMsg)
			return
		}
		for _, o := range listOrders.Orders {
			t.Log(o)
		}
	})
	t.Run("获取商品列表_空", func(t *testing.T) {
		// 测试空数据
		listOrders, err := clients.Order.ListOrders(context.Background(), &order.ListOrdersRequest{
			Pagination: &order.ListOrdersRequest_Pagination{
				Page:     100,
				PageSize: 10,
			},
			UserId: user.UserID,
		})
		if err != nil {
			t.Error(err)
		}
		if listOrders.StatusCode != code.Success {
			t.Log(listOrders.StatusMsg)
			return
		}
		for _, o := range listOrders.Orders {
			t.Log(o)
		}
	})

}
func TestGetOrder(t *testing.T) {
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

	t.Run("获取订单详情", func(t *testing.T) {
		orderDetail, err := clients.Order.GetOrder(context.Background(), &order.GetOrderRequest{
			OrderId: createResp.GetOrder().GetOrderId(),
			UserId:  user.UserID,
		})
		if err != nil {
			t.Error(err)
		}
		if orderDetail.StatusCode != code.Success {
			t.Log(orderDetail.StatusMsg)
			return
		}
		t.Logf("orderDetail: %+v", orderDetail.Order)
		t.Logf("items: %+v", orderDetail.Items)
		t.Logf("addres: %+v", orderDetail.Address)
	})
	t.Run("订单不存在", func(t *testing.T) {
		// 测试空数据
		orderDetail, err := clients.Order.GetOrder(context.Background(), &order.GetOrderRequest{
			OrderId: "0aaacb632c4aa",
			UserId:  user.UserID,
		})
		if err != nil {
			t.Error(err)
		}
		if orderDetail.StatusCode != code.OrderNotExist {
			t.Log(orderDetail.StatusMsg)
			return
		}
	})
	t.Run("订单内部接口调用", func(t *testing.T) {
		orderDetail, err := clients.Order.GetOrder2Payment(context.Background(), &order.GetOrderRequest{
			OrderId: createResp.GetOrder().GetOrderId(),
			UserId:  user.UserID,
		})
		if err != nil {
			t.Error(err)
		}
		if orderDetail.StatusCode != code.Success {
			t.Log(orderDetail.StatusMsg)
			return
		}
		t.Logf("orderDetail: %+v", orderDetail.Order)
	})

}
