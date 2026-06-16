package gateway_smoke

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"testing"
	"time"

	checkoutpb "github.com/falconfan123/Go-mall/services/checkout/pb"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/gatewayhttp"
	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"github.com/stretchr/testify/require"
)

type loginResp struct {
	StatusCode         int    `json:"statusCode"`
	StatusCodeLegacy   int    `json:"status_code"`
	StatusMsg          string `json:"statusMsg"`
	StatusMsgLegacy    string `json:"status_msg"`
	UserID             uint32 `json:"userId"`
	UserIDLegacy       uint32 `json:"user_id"`
	AccessToken        string `json:"accessToken"`
	AccessTokenLegacy  string `json:"access_token"`
	RefreshToken       string `json:"refreshToken"`
	RefreshTokenLegacy string `json:"refresh_token"`
	ShortToken         string `json:"shortToken"`
	ShortTokenLegacy   string `json:"short_token"`
	LongToken          string `json:"longToken"`
	LongTokenLegacy    string `json:"long_token"`
}

type getUserResp struct {
	StatusCode       int    `json:"statusCode"`
	StatusCodeLegacy int    `json:"status_code"`
	StatusMsg        string `json:"statusMsg"`
	StatusMsgLegacy  string `json:"status_msg"`
	UserID           uint32 `json:"userId"`
	UserIDLegacy     uint32 `json:"user_id"`
	UserName         string `json:"userName"`
	UserNameLegacy   string `json:"user_name"`
	Email            string `json:"email"`
}

type httpProduct struct {
	Id uint32 `json:"id"`
}

type listProductsResp struct {
	StatusCode       int           `json:"statusCode"`
	StatusCodeLegacy int           `json:"status_code"`
	StatusMsg        string        `json:"statusMsg"`
	StatusMsgLegacy  string        `json:"status_msg"`
	Products         []httpProduct `json:"products"`
}

type getProductResp struct {
	StatusCode       int          `json:"statusCode"`
	StatusCodeLegacy int          `json:"status_code"`
	StatusMsg        string       `json:"statusMsg"`
	StatusMsgLegacy  string       `json:"status_msg"`
	Product          *httpProduct `json:"product"`
}

type cartListResp struct {
	StatusCode       int    `json:"statusCode"`
	StatusCodeLegacy int    `json:"status_code"`
	StatusMsg        string `json:"statusMsg"`
	StatusMsgLegacy  string `json:"status_msg"`
	Data             []struct {
		ProductID          int32   `json:"productId"`
		ProductIDLegacy    int32   `json:"product_id"`
		ProductName        string  `json:"productName"`
		ProductNameLegacy  string  `json:"product_name"`
		ProductPrice       float64 `json:"productPrice"`
		ProductPriceLegacy float64 `json:"product_price"`
		Quantity           int32   `json:"quantity"`
	} `json:"data"`
}

type addAddressResp struct {
	StatusCode       int    `json:"statusCode"`
	StatusCodeLegacy int    `json:"status_code"`
	StatusMsg        string `json:"statusMsg"`
	StatusMsgLegacy  string `json:"status_msg"`
	Data             *struct {
		AddressID       uint64 `json:"addressId,string"`
		AddressIDLegacy uint64 `json:"address_id,string"`
	} `json:"data"`
}

type checkoutResp struct {
	StatusCode       int    `json:"statusCode"`
	StatusCodeLegacy int    `json:"status_code"`
	StatusMsg        string `json:"statusMsg"`
	StatusMsgLegacy  string `json:"status_msg"`
	PreOrderID       string `json:"preOrderId"`
	PreOrderIDLegacy string `json:"pre_order_id"`
}

type createOrderResp struct {
	StatusCode       int    `json:"statusCode"`
	StatusCodeLegacy int    `json:"status_code"`
	StatusMsg        string `json:"statusMsg"`
	StatusMsgLegacy  string `json:"status_msg"`
	Order            *struct {
		OrderID          string `json:"orderId"`
		OrderIDLegacy    string `json:"order_id"`
		PreOrderID       string `json:"preOrderId"`
		PreOrderIDLegacy string `json:"pre_order_id"`
		UserID           uint32 `json:"userId"`
		UserIDLegacy     uint32 `json:"user_id"`
	} `json:"order"`
}

type listOrdersResp struct {
	StatusCode       int    `json:"statusCode"`
	StatusCodeLegacy int    `json:"status_code"`
	StatusMsg        string `json:"statusMsg"`
	StatusMsgLegacy  string `json:"status_msg"`
	Orders           []struct {
		OrderID       string `json:"orderId"`
		OrderIDLegacy string `json:"order_id"`
		OrderStatus   int32  `json:"order_status"`
	} `json:"orders"`
}

type paymentResp struct {
	StatusCode       int    `json:"statusCode"`
	StatusCodeLegacy int    `json:"status_code"`
	StatusMsg        string `json:"statusMsg"`
	StatusMsgLegacy  string `json:"status_msg"`
	Payment          *struct {
		OrderID       string `json:"orderId"`
		OrderIDLegacy string `json:"order_id"`
	} `json:"payment"`
}

type tokenResp struct {
	PathKey         string `json:"pathKey"`
	PathKeyLegacy   string `json:"path_key"`
	ExpiresAt       int64  `json:"expiresAt,string"`
	ExpiresAtLegacy int64  `json:"expires_at,string"`
}

type seckillResp struct {
	StatusCode       int    `json:"statusCode"`
	StatusCodeLegacy int    `json:"status_code"`
	StatusMsg        string `json:"statusMsg"`
	StatusMsgLegacy  string `json:"status_msg"`
	OrderID          string `json:"orderId"`
	OrderIDLegacy    string `json:"order_id"`
	Message          string `json:"message"`
}

func pickInt(primary, legacy int) int {
	if primary != 0 {
		return primary
	}
	return legacy
}

func pickUint32(primary, legacy uint32) uint32 {
	if primary != 0 {
		return primary
	}
	return legacy
}

func pickUint64(primary, legacy uint64) uint64 {
	if primary != 0 {
		return primary
	}
	return legacy
}

func pickInt32(primary, legacy int32) int32 {
	if primary != 0 {
		return primary
	}
	return legacy
}

func pickInt64(primary, legacy int64) int64 {
	if primary != 0 {
		return primary
	}
	return legacy
}

func pickString(primary, legacy string) string {
	if primary != "" {
		return primary
	}
	return legacy
}

func TestGatewayHTTPHappyPath(t *testing.T) {
	clients := harness.NewClients(t)
	harness.WaitForServices(t, clients)

	user := seed.CreateUser(t, clients.Users)
	product := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 20)

	gateway := gatewayhttp.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), testenv.Timeout())
	defer cancel()

	var login loginResp
	resp, body, err := gateway.DoJSON(ctx, "POST", "/api/v1/users/login", nil, map[string]any{
		"username": user.Username,
		"password": user.Password,
		"ip":       "127.0.0.1",
	}, &login)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(login.StatusCode, login.StatusCodeLegacy), string(body))
	accessToken := pickString(
		pickString(login.AccessToken, login.AccessTokenLegacy),
		pickString(login.ShortToken, login.ShortTokenLegacy),
	)
	refreshToken := pickString(
		pickString(login.RefreshToken, login.RefreshTokenLegacy),
		pickString(login.LongToken, login.LongTokenLegacy),
	)
	require.NotEmpty(t, accessToken, string(body))
	require.NotEmpty(t, refreshToken, string(body))

	authGateway := gateway.WithTokens(
		accessToken,
		refreshToken,
	).WithUserID(user.UserID)

	var me getUserResp
	resp, body, err = authGateway.DoJSON(ctx, "GET", "/api/v1/users/me", map[string]string{
		"user_id": strconv.FormatUint(uint64(user.UserID), 10),
	}, nil, &me)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(me.StatusCode, me.StatusCodeLegacy), string(body))
	require.Equal(t, user.UserID, pickUint32(me.UserID, me.UserIDLegacy), string(body))

	var productList listProductsResp
	resp, body, err = authGateway.DoJSON(ctx, "GET", "/api/v1/products", map[string]string{
		"cursor": "0",
		"limit":  "100",
	}, nil, &productList)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(productList.StatusCode, productList.StatusCodeLegacy), string(body))
	require.NotEmpty(t, productList.Products)

	var productDetail getProductResp
	resp, body, err = authGateway.DoJSON(ctx, "GET", "/api/v1/products/detail", map[string]string{
		"id": strconv.FormatInt(product.ProductID, 10),
	}, nil, &productDetail)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(productDetail.StatusCode, productDetail.StatusCodeLegacy), string(body))
	require.NotNil(t, productDetail.Product)
	require.Equal(t, uint32(product.ProductID), productDetail.Product.Id)

	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/cart/items", nil, map[string]any{
		"user_id":       user.UserID,
		"product_id":    product.ProductID,
		"product_name":  product.Name,
		"product_image": "",
		"productPrice":  float64(product.Price) / 100,
		"quantity":      0,
		"checked":       true,
	}, nil)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)

	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/cart/items", nil, map[string]any{
		"user_id":       user.UserID,
		"product_id":    product.ProductID,
		"product_name":  product.Name,
		"product_image": "",
		"productPrice":  float64(product.Price) / 100,
		"quantity":      0,
		"checked":       true,
	}, nil)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)

	var cartList cartListResp
	resp, body, err = authGateway.DoJSON(ctx, "GET", "/api/v1/cart/items", map[string]string{
		"id":      strconv.FormatUint(uint64(user.UserID), 10),
		"user_id": strconv.FormatUint(uint64(user.UserID), 10),
	}, nil, &cartList)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(cartList.StatusCode, cartList.StatusCodeLegacy), string(body))
	require.Len(t, cartList.Data, 1)
	require.Equal(t, int32(product.ProductID), pickInt32(cartList.Data[0].ProductID, cartList.Data[0].ProductIDLegacy), string(body))
	require.Equal(t, int32(2), cartList.Data[0].Quantity)

	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/cart/items/sub", nil, map[string]any{
		"user_id":    user.UserID,
		"product_id": product.ProductID,
	}, nil)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)

	resp, body, err = authGateway.DoJSON(ctx, "GET", "/api/v1/cart/items", map[string]string{
		"id":      strconv.FormatUint(uint64(user.UserID), 10),
		"user_id": strconv.FormatUint(uint64(user.UserID), 10),
	}, nil, &cartList)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Len(t, cartList.Data, 1)
	require.Equal(t, int32(1), cartList.Data[0].Quantity)

	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/cart/items/delete", nil, map[string]any{
		"user_id":    user.UserID,
		"product_id": product.ProductID,
	}, nil)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)

	resp, body, err = authGateway.DoJSON(ctx, "GET", "/api/v1/cart/items", map[string]string{
		"id":      strconv.FormatUint(uint64(user.UserID), 10),
		"user_id": strconv.FormatUint(uint64(user.UserID), 10),
	}, nil, &cartList)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Empty(t, cartList.Data)

	var newAddress addAddressResp
	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/users/addresses", nil, map[string]any{
		"user_id":          user.UserID,
		"recipient_name":   "gateway-smoke",
		"phone_number":     "13800000001",
		"province":         "Shanghai",
		"city":             "Shanghai",
		"detailed_address": "gateway smoke address",
		"is_default":       false,
		"ip":               "127.0.0.1",
	}, &newAddress)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(newAddress.StatusCode, newAddress.StatusCodeLegacy), string(body))
	require.NotNil(t, newAddress.Data)

	checkoutProduct := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 20)

	var prepare checkoutResp
	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/checkout/prepare", nil, map[string]any{
		"user_id":    user.UserID,
		"address_id": pickUint64(newAddress.Data.AddressID, newAddress.Data.AddressIDLegacy),
		"coupon_id":  "",
		"order_items": []map[string]any{
			{
				"product_id": checkoutProduct.ProductID,
				"quantity":   1,
			},
		},
	}, &prepare)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(prepare.StatusCode, prepare.StatusCodeLegacy), string(body))
	require.NotEmpty(t, pickString(prepare.PreOrderID, prepare.PreOrderIDLegacy), string(body))

	orderID := testenv.UniqueName("gw-order")
	var createdOrder createOrderResp
	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/orders", nil, map[string]any{
		"pre_order_id":   pickString(prepare.PreOrderID, prepare.PreOrderIDLegacy),
		"user_id":        user.UserID,
		"coupon_id":      "",
		"address_id":     pickUint64(newAddress.Data.AddressID, newAddress.Data.AddressIDLegacy),
		"payment_method": int(orderpb.PaymentMethod_ALIPAY),
		"order_id":       orderID,
	}, &createdOrder)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(createdOrder.StatusCode, createdOrder.StatusCodeLegacy), string(body))
	require.NotNil(t, createdOrder.Order)
	require.Equal(t, orderID, pickString(createdOrder.Order.OrderID, createdOrder.Order.OrderIDLegacy), string(body))

	var orderList listOrdersResp
	resp, body, err = authGateway.DoJSON(ctx, "GET", "/api/v1/orders", map[string]string{
		"user_id": strconv.FormatUint(uint64(user.UserID), 10),
	}, nil, &orderList)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Contains(t, string(body), orderID)

	var createdPayment paymentResp
	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/payments", nil, map[string]any{
		"user_id":        user.UserID,
		"order_id":       orderID,
		"payment_method": 3,
	}, &createdPayment)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(createdPayment.StatusCode, createdPayment.StatusCodeLegacy), string(body))
	require.NotNil(t, createdPayment.Payment)
	require.Equal(t, orderID, pickString(createdPayment.Payment.OrderID, createdPayment.Payment.OrderIDLegacy), string(body))

	cancelPrepare, err := clients.Checkout.PrepareCheckout(ctx, seed.MakeCheckoutRequest(
		user.UserID,
		user.AddressID,
		&checkoutpb.CheckoutReq_OrderItem{ProductId: int32(checkoutProduct.ProductID), Quantity: 1},
	))
	require.NoError(t, err)

	cancelOrderID := testenv.UniqueName("gw-cancel")
	_, err = clients.Order.CreateOrder(ctx, seed.MakeOrderRequest(
		cancelPrepare.GetPreOrderId(),
		user.UserID,
		user.AddressID,
		orderpb.PaymentMethod_ALIPAY,
		cancelOrderID,
	))
	require.NoError(t, err)

	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/orders/cancel", nil, map[string]any{
		"order_id":      cancelOrderID,
		"user_id":       user.UserID,
		"cancel_reason": "gateway smoke",
		"initiative":    true,
	}, nil)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)

	seckillProduct := seed.CreateProductWithInventory(t, clients.Product, clients.Inventory, 50)
	configureSeckillActivity(t, seckillProduct.ProductID, 10)

	var token tokenResp
	resp, body, err = authGateway.DoJSON(ctx, "GET", "/api/v1/activity/token", map[string]string{
		"activity_id": strconv.FormatInt(seckillProduct.ProductID, 10),
	}, nil, &token)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.NotEmpty(t, pickString(token.PathKey, token.PathKeyLegacy), string(body))

	var seckill seckillResp
	resp, body, err = authGateway.DoJSON(ctx, "POST", "/api/v1/orders/seckill", nil, map[string]any{
		"path_key":   pickString(token.PathKey, token.PathKeyLegacy),
		"product_id": seckillProduct.ProductID,
		"user_id":    user.UserID,
		"address_id": user.AddressID,
	}, &seckill)
	require.NoError(t, err)
	gatewayhttp.RequireStatusOK(t, resp, body)
	require.Equal(t, 0, pickInt(seckill.StatusCode, seckill.StatusCodeLegacy), string(body))
	require.NotEmpty(t, pickString(seckill.OrderID, seckill.OrderIDLegacy), string(body))
}

func configureSeckillActivity(t *testing.T, productID int64, stock int64) {
	t.Helper()
	now := time.Now().Add(-2 * time.Second).UnixMilli()

	setRedisKey(t, "act_start_limit", fmt.Sprintf("%d", now))
	setRedisKey(t, fmt.Sprintf("act_%d_stock", productID), fmt.Sprintf("%d", stock))
	deleteRedisKey(t, fmt.Sprintf("act_%d_bought", productID))
}

func setRedisKey(t *testing.T, key, value string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "exec", "go-mall-redis", "redis-cli", "SET", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("set redis key failed key=%s err=%v output=%s", key, err, string(output))
	}
}

func deleteRedisKey(t *testing.T, key string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "exec", "go-mall-redis", "redis-cli", "DEL", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("delete redis key failed key=%s err=%v output=%s", key, err, string(output))
	}
}
