package svc

import (
	"github.com/falconfan123/Go-mall/apis/coupon/internal/config"
	"github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/coupons/pb"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                config.Config
	CouponRpc             couponsclient.Coupons
	WithClientMiddleware  rest.Middleware
	WrapperAuthMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		CouponRpc:             couponsclient.NewCoupons(zrpc.MustNewClient(c.CouponRpc)),
		WithClientMiddleware:  middleware.WithClientMiddleware,
		WrapperAuthMiddleware: middleware.WrapperAuthMiddleware(c.AuthsRpc, c.WhitePathList, c.OptionPathList),
	}
}
