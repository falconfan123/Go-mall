package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

// Config is the configuration struct for the service.
type Config struct {
	rest.RestConf

	AuthsRPC zrpc.RpcClientConf

	UserRPC        zrpc.RpcClientConf
	Consul         consul.Conf
	WhitePathList  []string
	OptionPathList []string
}
