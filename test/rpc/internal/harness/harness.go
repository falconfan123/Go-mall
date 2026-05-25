package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	auditpb "github.com/falconfan123/Go-mall/services/audit/pb"
	authspb "github.com/falconfan123/Go-mall/services/auths/pb"
	cartspb "github.com/falconfan123/Go-mall/services/carts/pb"
	checkoutpb "github.com/falconfan123/Go-mall/services/checkout/pb"
	couponspb "github.com/falconfan123/Go-mall/services/coupons/pb"
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

func WaitForStackReady(ctx context.Context) error {
	for _, check := range stackChecks() {
		if err := waitForRPC(ctx, check.name, check.fn); err != nil {
			return err
		}
	}

	return nil
}

func MonitorStackHealth(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		for _, check := range stackChecks() {
			checkCtx, cancel := context.WithTimeout(ctx, interval)
			err := check.fn(checkCtx)
			cancel()
			if err != nil {
				return enrichStackError(check.name, err)
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

type stackCheck struct {
	name string
	fn   func(context.Context) error
}

func stackChecks() []stackCheck {
	return []stackCheck{
		{
			name: "auths",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "auths", biz.AuthsRpcPort, func(conn *grpc.ClientConn) error {
					_, err := authspb.NewAuthsClient(conn).GenerateToken(ctx, &authspb.AuthGenReq{
						UserId:   1,
						Username: "readycheck",
						ClientIp: "127.0.0.1",
					})
					return err
				})
			},
		},
		{
			name: "users",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "users", biz.UsersRpcPort, func(conn *grpc.ClientConn) error {
					_, err := userspb.NewUsersClient(conn).Ping(ctx, &userspb.Request{})
					return err
				})
			},
		},
		{
			name: "product",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "product", biz.ProductRpcPort, func(conn *grpc.ClientConn) error {
					_, err := productpb.NewProductCatalogServiceClient(conn).GetAllProduct(ctx, &productpb.GetAllProductsReq{
						Page:     1,
						PageSize: 1,
					})
					return err
				})
			},
		},
		{
			name: "carts",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "carts", biz.CartsRpcPort, func(conn *grpc.ClientConn) error {
					_, err := cartspb.NewCartClient(conn).CartItemList(ctx, &cartspb.UserInfo{Id: 0})
					return err
				})
			},
		},
		{
			name: "checkout",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "checkout", biz.CheckoutRpcPort, func(conn *grpc.ClientConn) error {
					_, err := checkoutpb.NewCheckoutServiceClient(conn).GetCheckoutList(ctx, &checkoutpb.CheckoutListReq{
						UserId:   0,
						Page:     1,
						PageSize: 1,
					})
					return err
				})
			},
		},
		{
			name: "inventory",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "inventory", biz.InventoryRpcPort, func(conn *grpc.ClientConn) error {
					_, err := inventorypb.NewInventoryClient(conn).GetInventory(ctx, &inventorypb.GetInventoryReq{ProductId: 0})
					return err
				})
			},
		},
		{
			name: "order",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "order", biz.OrderRpcPort, func(conn *grpc.ClientConn) error {
					_, err := orderpb.NewOrderServiceClient(conn).ListOrders(ctx, &orderpb.ListOrdersRequest{
						UserId: 0,
						Pagination: &orderpb.ListOrdersRequest_Pagination{
							Page:     1,
							PageSize: 1,
						},
					})
					return err
				})
			},
		},
		{
			name: "payment",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "payment", biz.PaymentRpcPort, func(conn *grpc.ClientConn) error {
					_, err := paymentpb.NewPaymentClient(conn).ListPayments(ctx, &paymentpb.PaymentListReq{
						UserId: 0,
						Pagination: &paymentpb.PaymentListReq_Pagination{
							Page:     1,
							PageSize: 1,
						},
					})
					return err
				})
			},
		},
		{
			name: "audit",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "audit", biz.AuditRpcPort, func(conn *grpc.ClientConn) error {
					_, err := auditpb.NewAuditClient(conn).CreateAuditLog(ctx, &auditpb.CreateAuditLogReq{
						UserId:      1,
						ActionType:  "READYCHECK",
						TargetTable: "health",
						TargetId:    1,
						ServiceName: "readycheck",
					})
					return err
				})
			},
		},
		{
			name: "coupons",
			fn: func(ctx context.Context) error {
				return withConn(ctx, "coupons", biz.CouponsRpcPort, func(conn *grpc.ClientConn) error {
					_, err := couponspb.NewCouponsClient(conn).ListCoupons(ctx, &couponspb.ListCouponsReq{
						Pagination: &couponspb.PaginationReq{
							Page: 1,
							Size: 1,
						},
					})
					return err
				})
			},
		},
	}
}

type stackInspectRecord struct {
	Name        string `json:"name"`
	Pid         int    `json:"pid"`
	Port        int    `json:"port"`
	Alive       bool   `json:"alive"`
	Listening   bool   `json:"listening"`
	ExitCode    string `json:"exitCode"`
	FinishedAt  string `json:"finishedAt"`
	LogPath     string `json:"logPath"`
	LastLogLine string `json:"lastLogLine"`
}

func enrichStackError(service string, err error) error {
	records, inspectErr := inspectLocalStack()
	if inspectErr != nil {
		return fmt.Errorf("%s: %w (stack inspect failed: %v)", service, err, inspectErr)
	}
	for _, record := range records {
		if record.Name != service {
			continue
		}
		return fmt.Errorf(
			"%s: %w (pid=%d port=%d alive=%t listening=%t exit_code=%s finished_at=%s last_log=%q)",
			service,
			err,
			record.Pid,
			record.Port,
			record.Alive,
			record.Listening,
			emptyFallback(record.ExitCode, "unknown"),
			emptyFallback(record.FinishedAt, "n/a"),
			record.LastLogLine,
		)
	}
	return fmt.Errorf("%s: %w", service, err)
}

func inspectLocalStack() ([]stackInspectRecord, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("bash", filepath.Join(root, "scripts/ci-rpc-stack.sh"), "inspect", "--json")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var records []stackInspectRecord
	if err := json.Unmarshal(out, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(trimTrailingWhitespace(out)), nil
}

func trimTrailingWhitespace(b []byte) []byte {
	for len(b) > 0 {
		switch b[len(b)-1] {
		case '\n', '\r', '\t', ' ':
			b = b[:len(b)-1]
		default:
			return b
		}
	}
	return b
}

func emptyFallback(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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

func withConn(ctx context.Context, service string, port int, fn func(*grpc.ClientConn) error) error {
	conn, err := dialContext(ctx, service, port)
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(conn)
}

func dialContext(ctx context.Context, service string, port int) (*grpc.ClientConn, error) {
	return grpc.DialContext(
		ctx,
		testenv.ServiceAddr(service, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

func mustDial(t *testing.T, service string, port int) *grpc.ClientConn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := dialContext(ctx, service, port)
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
