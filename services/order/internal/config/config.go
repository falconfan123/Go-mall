package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	RedisConf      redis.RedisConf
	Consul         consul.Conf

	InventoryRpc zrpc.RpcClientConf
	CouponsRpc   zrpc.RpcClientConf
	ProductRpc   zrpc.RpcClientConf
	Prometheus   PrometheusConf
}

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}

type PrometheusConf struct {
	Host string
	Port int
	Path string
}
