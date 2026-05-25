package harness

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	checkoutpb "github.com/falconfan123/Go-mall/services/checkout/pb"
	inventorypb "github.com/falconfan123/Go-mall/services/inventory/pb"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	paymentpb "github.com/falconfan123/Go-mall/services/payment/pb"
	productpb "github.com/falconfan123/Go-mall/services/product/pb"
	userspb "github.com/falconfan123/Go-mall/services/users/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Clients struct {
	Users     userspb.UsersClient
	Product   productpb.ProductCatalogServiceClient
	Inventory inventorypb.InventoryClient
	Checkout  checkoutpb.CheckoutServiceClient
	Order     orderpb.OrderServiceClient
	Payment   paymentpb.PaymentClient
}

func NewClients(t *testing.T) *Clients {
	t.Helper()
	return &Clients{
		Users:     mustUsersClient(t),
		Product:   mustProductClient(t),
		Inventory: mustInventoryClient(t),
		Checkout:  mustCheckoutClient(t),
		Order:     mustOrderClient(t),
		Payment:   mustPaymentClient(t),
	}
}

func WaitForServices(t *testing.T, clients *Clients) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testenv.Timeout())
	defer cancel()

	wait := func(name string, fn func(context.Context) error) {
		t.Helper()
		if err := waitForRPC(ctx, name, fn); err != nil {
			t.Fatalf("service %s not ready: %v", name, err)
		}
	}

	wait("users", func(ctx context.Context) error {
		_, err := clients.Users.Ping(ctx, &userspb.Request{})
		return err
	})
	wait("product", func(ctx context.Context) error {
		_, err := clients.Product.GetAllProduct(ctx, &productpb.GetAllProductsReq{Page: 1, PageSize: 1})
		return err
	})
	wait("inventory", func(ctx context.Context) error {
		_, err := clients.Inventory.GetInventory(ctx, &inventorypb.GetInventoryReq{ProductId: 0})
		return err
	})
	wait("checkout", func(ctx context.Context) error {
		_, err := clients.Checkout.GetCheckoutList(ctx, &checkoutpb.CheckoutListReq{UserId: 0, Page: 1, PageSize: 1})
		return err
	})
	wait("order", func(ctx context.Context) error {
		_, err := clients.Order.ListOrders(ctx, &orderpb.ListOrdersRequest{UserId: 0, Pagination: &orderpb.ListOrdersRequest_Pagination{Page: 1, PageSize: 1}})
		return err
	})
	wait("payment", func(ctx context.Context) error {
		_, err := clients.Payment.ListPayments(ctx, &paymentpb.PaymentListReq{UserId: 0, Pagination: &paymentpb.PaymentListReq_Pagination{Page: 1, PageSize: 1}})
		return err
	})
}

func waitForRPC(ctx context.Context, name string, fn func(context.Context) error) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error
	for {
		if err := fn(ctx); err == nil {
			return nil
		} else {
			lastErr = err
			switch status.Code(err) {
			case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
			default:
				return nil
			}
		}

		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("%s: %w", name, lastErr)
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func mustDial(t *testing.T, service string, port int) *grpc.ClientConn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		testenv.ServiceAddr(service, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("connect %s failed: %v", service, err)
	}
	return conn
}

func mustUsersClient(t *testing.T) userspb.UsersClient {
	return userspb.NewUsersClient(mustDial(t, "users", biz.UsersRpcPort))
}

func mustProductClient(t *testing.T) productpb.ProductCatalogServiceClient {
	return productpb.NewProductCatalogServiceClient(mustDial(t, "product", biz.ProductRpcPort))
}

func mustInventoryClient(t *testing.T) inventorypb.InventoryClient {
	return inventorypb.NewInventoryClient(mustDial(t, "inventory", biz.InventoryRpcPort))
}

func mustCheckoutClient(t *testing.T) checkoutpb.CheckoutServiceClient {
	return checkoutpb.NewCheckoutServiceClient(mustDial(t, "checkout", biz.CheckoutRpcPort))
}

func mustOrderClient(t *testing.T) orderpb.OrderServiceClient {
	return orderpb.NewOrderServiceClient(mustDial(t, "order", biz.OrderRpcPort))
}

func mustPaymentClient(t *testing.T) paymentpb.PaymentClient {
	return paymentpb.NewPaymentClient(mustDial(t, "payment", biz.PaymentRpcPort))
}
