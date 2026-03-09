package svc

import (
	"jijizhazha1024/go-mall/apis/flash_sale/internal/config"
	"jijizhazha1024/go-mall/common/middleware"
	"jijizhazha1024/go-mall/services/checkout/checkoutservice"
	"jijizhazha1024/go-mall/services/order/orderservice"
	"jijizhazha1024/go-mall/services/users/usersclient"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                config.Config
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
	CheckoutRpc           checkoutservice.CheckoutService
	OrderRpc              orderservice.OrderService
	UsersRpc              usersclient.Users
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		WithClientMiddleware:  middleware.WithClientMiddleware,
		WrapperAuthMiddleware: middleware.WrapperAuthMiddleware(c.AuthsRpc, nil, c.OptionPathList),
		CheckoutRpc:           checkoutservice.NewCheckoutService(zrpc.MustNewClient(c.CheckoutRpc)),
		OrderRpc:              orderservice.NewOrderService(zrpc.MustNewClient(c.OrderRpc)),
		UsersRpc:              usersclient.NewUsers(zrpc.MustNewClient(c.UsersRpc)),
	}
}
