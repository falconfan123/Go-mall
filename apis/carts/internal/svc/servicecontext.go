package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/falconfan123/Go-mall/apis/carts/internal/config"
	"github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/carts/cartsclient"
	"github.com/falconfan123/Go-mall/services/product/product"
)

type ServiceContext struct {
	Config                config.Config
	CartsRpc              cartsclient.Cart
	ProductRpc            product.ProductCatalogService
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		CartsRpc:              cartsclient.NewCart(zrpc.MustNewClient(c.CartsRpc)),
		ProductRpc:            product.NewProductCatalogService(zrpc.MustNewClient(c.ProductRpc)),
		WrapperAuthMiddleware: middleware.WrapperAuthMiddleware(c.AuthsRpc, nil, nil),
		WithClientMiddleware:  middleware.WithClientMiddleware,
	}
}
