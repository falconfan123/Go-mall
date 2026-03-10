package svc

import (
	"github.com/falconfan123/Go-mall/apis/user/internal/config"
	"github.com/falconfan123/Go-mall/common/middleware"
	"github.com/falconfan123/Go-mall/services/auths/authsclient"
	"github.com/falconfan123/Go-mall/services/users/usersclient"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                config.Config
	UserRpc               usersclient.Users
	AuthsRpc              authsclient.Auths
	WrapperAuthMiddleware rest.Middleware
	WithClientMiddleware  rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		UserRpc:               usersclient.NewUsers(zrpc.MustNewClient(c.UserRpc)),
		AuthsRpc:              authsclient.NewAuths(zrpc.MustNewClient(c.AuthsRpc)),
		Config:                c,
		WrapperAuthMiddleware: middleware.WrapperAuthMiddleware(c.AuthsRpc, c.WhitePathList, c.OptionPathList), // # 需要指定认证rpc地址

		WithClientMiddleware: middleware.WithClientMiddleware,
	}
}
