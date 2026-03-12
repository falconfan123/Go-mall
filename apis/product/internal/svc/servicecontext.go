package svc

import (
	"github.com/falconfan123/Go-mall/apis/product/internal/config"
	"github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/product/product"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                config.Config
	ProductRpc            product.ProductCatalogService
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		ProductRpc:            product.NewProductCatalogService(zrpc.MustNewClient(c.ProductRpc)),
		WithClientMiddleware:  middleware.WithClientMiddleware,
		WrapperAuthMiddleware: middleware.WrapperAuthMiddleware(c.AuthsRpc, nil, c.OptionPathList),
	}
}
