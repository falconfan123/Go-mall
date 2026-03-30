package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf
	Consul     consul.Conf
	Prometheus PrometheusConf
}

type PrometheusConf struct {
	Host string
	Port int
	Path string
}
