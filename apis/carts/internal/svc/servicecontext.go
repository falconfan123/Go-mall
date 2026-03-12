package svc

import (
	"github.com/falconfan123/Go-mall/apis/carts/internal/config"
	"github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/carts/cartsclient"
	"github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext is the service context containing dependencies.
type ServiceContext struct {
	Config                config.Config
	CartsRPC              cartsclient.Cart
	ProductRPC            product.ProductCatalogService
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
}

// NewServiceContext creates a new service context.
func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		CartsRPC:              cartsclient.NewCart(zrpc.MustNewClient(c.CartsRPC)),
		ProductRPC:            product.NewProductCatalogService(zrpc.MustNewClient(c.ProductRPC)),
		WrapperAuthMiddleware: middleware.WrapperAuthMiddleware(c.AuthsRPC, nil, nil),
		WithClientMiddleware:  middleware.WithClientMiddleware,
	}
}
