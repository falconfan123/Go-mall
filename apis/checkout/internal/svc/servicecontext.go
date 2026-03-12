package svc

import (
	"github.com/falconfan123/Go-mall/apis/checkout/internal/config"
	"github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext is the service context containing dependencies.
type ServiceContext struct {
	Config                config.Config
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
	CheckoutRPC           checkoutservice.CheckoutService
}

// NewServiceContext creates a new service context.
func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		WithClientMiddleware:  middleware.WithClientMiddleware,
		WrapperAuthMiddleware: middleware.WrapperAuthMiddleware(c.AuthsRPC, nil, nil),
		CheckoutRPC:           checkoutservice.NewCheckoutService(zrpc.MustNewClient(c.CheckoutRPC)),
	}
}
