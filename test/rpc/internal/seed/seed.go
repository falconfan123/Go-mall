package seed

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	_ "github.com/lib/pq"

	checkoutpb "github.com/falconfan123/Go-mall/services/checkout/pb"
	inventorypb "github.com/falconfan123/Go-mall/services/inventory/pb"
	orderspb "github.com/falconfan123/Go-mall/services/order/pb"
	productpb "github.com/falconfan123/Go-mall/services/product/pb"
	userspb "github.com/falconfan123/Go-mall/services/users/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
)

var seedCounter uint64

type UserSeed struct {
	UserID    uint32
	AddressID uint64
	Username  string
	Email     string
	Password  string
}

type ProductSeed struct {
	ProductID int64
	Price     int64
	Name      string
}

func CreateUser(t *testing.T, client userspb.UsersClient) UserSeed {
	t.Helper()
	idx := atomic.AddUint64(&seedCounter, 1)
	username := testenv.UniqueName(fmt.Sprintf("user-%d", idx))
	email := fmt.Sprintf("%s@example.com", username)
	password := "Password123!"
	resp, err := client.Register(context.Background(), &userspb.RegisterRequest{
		Username:        username,
		Email:           email,
		Password:        password,
		ConfirmPassword: password,
		Ip:              "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("register user failed: %v", err)
	}
	if resp.GetStatusCode() != 0 {
		t.Fatalf("register user rejected: %s", resp.GetStatusMsg())
	}

	addrResp, err := client.AddAddress(context.Background(), &userspb.AddAddressRequest{
		RecipientName:   username,
		PhoneNumber:     "13800000000",
		Province:        "Shanghai",
		City:            "Shanghai",
		DetailedAddress: "integration test",
		IsDefault:       true,
		UserId:          resp.GetUserId(),
		Ip:              "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("add address failed: %v", err)
	}
	if addrResp.GetStatusCode() != 0 || addrResp.GetData() == nil {
		t.Fatalf("add address rejected: %s", addrResp.GetStatusMsg())
	}

	return UserSeed{
		UserID:    resp.GetUserId(),
		AddressID: addrResp.GetData().GetAddressId(),
		Username:  username,
		Email:     email,
		Password:  password,
	}
}

func CreateProductWithInventory(t *testing.T, productClient productpb.ProductCatalogServiceClient, inventoryClient inventorypb.InventoryClient, stock int64) ProductSeed {
	t.Helper()
	idx := atomic.AddUint64(&seedCounter, 1)
	name := testenv.UniqueName(fmt.Sprintf("product-%d", idx))
	price := int64(1999 + idx)
	resp, err := productClient.CreateProduct(context.Background(), &productpb.CreateProductReq{
		Name:        name,
		Description: "integration test product",
		Price:       price,
		Stock:       stock,
		Picture:     []byte("integration"),
		Categories:  []string{"1"},
	})
	if err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	if resp.GetStatusCode() != 0 {
		t.Fatalf("create product rejected: %s", resp.GetStatusMsg())
	}

	_, err = inventoryClient.UpdateInventory(context.Background(), &inventorypb.UpdateInventoryReq{
		Items: []*inventorypb.UpdateInventoryReq_Items{
			{ProductId: int32(resp.GetProductId()), Quantity: int32(stock)},
		},
	})
	if err != nil {
		t.Fatalf("seed inventory failed: %v", err)
	}

	return ProductSeed{
		ProductID: resp.GetProductId(),
		Price:     price,
		Name:      name,
	}
}

func MakeCheckoutRequest(userID uint32, addressID uint64, items ...*checkoutpb.CheckoutReq_OrderItem) *checkoutpb.CheckoutReq {
	return &checkoutpb.CheckoutReq{
		UserId:     userID,
		AddressId:  addressID,
		OrderItems: items,
	}
}

func MakeOrderRequest(preOrderID string, userID uint32, addressID uint64, paymentMethod orderspb.PaymentMethod, orderID string) *orderspb.CreateOrderRequest {
	return &orderspb.CreateOrderRequest{
		PreOrderId:    preOrderID,
		UserId:        userID,
		AddressId:     addressID,
		PaymentMethod: paymentMethod,
		OrderId:       orderID,
	}
}

// SeedUserCoupons 重置 coupons 集成测试所需的 user_coupons 种子行
// （user_id=1 持 5 张测试券；LOCK/RELEASE/USE/DUP=AVAILABLE(1)、USED=USED(3)）。
// coupons 测试会改变券状态（锁定/使用），故每个用例前重置到初始态。
// 直连 DB（集成栈 postgres），不改测试断言、不 skip 用例。
func SeedUserCoupons(t *testing.T) {
	t.Helper()
	host := os.Getenv("GO_MALL_CI_POSTGRES_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	dsn := fmt.Sprintf("postgres://root:fht3825099@%s:5432/mall?sslmode=disable", host)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("postgres unreachable for coupon seed: %v", err)
	}
	rows := []struct {
		couponID string
		status   int
	}{
		{"LOCK20250525001", 1},
		{"RELEASE20250525001", 1},
		{"USE20250525001", 1},
		{"USED20250525001", 3},
		{"DUP20250525001", 1},
	}
	for _, r := range rows {
		if _, err := db.Exec(`DELETE FROM user_coupons WHERE user_id=1 AND coupon_id=$1`, r.couponID); err != nil {
			t.Fatalf("clean user_coupon %s: %v", r.couponID, err)
		}
		if _, err := db.Exec(`INSERT INTO user_coupons (user_id, coupon_id, status) VALUES (1,$1,$2)`, r.couponID, r.status); err != nil {
			t.Fatalf("seed user_coupon %s: %v", r.couponID, err)
		}
	}
}
