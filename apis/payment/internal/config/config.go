package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config is the configuration struct for the service.
type Config struct {
	rest.RestConf
	AuthsRPC   zrpc.RpcClientConf
	PaymentRPC zrpc.RpcClientConf
}
