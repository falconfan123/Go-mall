package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	AuthsRpc       zrpc.RpcClientConf
	CheckoutRpc    zrpc.RpcClientConf
	OrderRpc       zrpc.RpcClientConf
	UsersRpc       zrpc.RpcClientConf
	OptionPathList []string `json:",optional"`
}
