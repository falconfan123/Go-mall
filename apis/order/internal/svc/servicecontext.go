package svc

import (
	"github.com/falconfan123/Go-mall/apis/order/internal/config"
	"github.com/falconfan123/Go-mall/apis/order/internal/middleware"
	commonmiddleware "github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext is the service context containing dependencies.
type ServiceContext struct {
	Config                config.Config
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
	OrderRPC              order.OrderServiceClient
}

// NewServiceContext creates a new service context.
func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		WithClientMiddleware:  commonmiddleware.WithClientMiddleware,
		WrapperAuthMiddleware: middleware.NewWrapperAuthMiddleware().Handle,
		OrderRPC:              order.NewOrderServiceClient(zrpc.MustNewClient(c.OrderRPC).Conn()),
	}
}
