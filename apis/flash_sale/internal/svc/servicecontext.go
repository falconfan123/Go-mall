package svc

import (
	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/config"
	"github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                config.Config
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
	CheckoutRpc           checkoutservice.CheckoutService
	OrderRpc              order.OrderServiceClient
	UsersRpc              users.Users
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		WithClientMiddleware:  middleware.WithClientMiddleware,
		WrapperAuthMiddleware: middleware.WrapperAuthMiddleware(c.AuthsRpc, nil, c.OptionPathList),
		CheckoutRpc:           checkoutservice.NewCheckoutService(zrpc.MustNewClient(c.CheckoutRpc)),
		OrderRpc:              order.NewOrderServiceClient(zrpc.MustNewClient(c.OrderRpc).Conn()),
		UsersRpc:              users.NewUsers(zrpc.MustNewClient(c.UsersRpc)),
	}
}
