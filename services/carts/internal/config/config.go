package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	Consul         consul.Conf
}
type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}
