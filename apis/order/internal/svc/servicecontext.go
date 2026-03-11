package svc

import (
	"github.com/falconfan123/Go-mall/apis/order/internal/config"
	"github.com/falconfan123/Go-mall/apis/order/internal/middleware"
	commonmiddleware "github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/order/order"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                config.Config
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
	OrderRpc              order.OrderServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		WithClientMiddleware:  commonmiddleware.WithClientMiddleware,
		WrapperAuthMiddleware: middleware.NewWrapperAuthMiddleware().Handle,
		OrderRpc:              order.NewOrderServiceClient(zrpc.MustNewClient(c.OrderRpc).Conn()),
	}
}
