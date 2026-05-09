package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	RedisConf      redis.RedisConf
	ProductRpc     zrpc.RpcClientConf
	PrometheusExt  PrometheusExtConf
}

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}

type PrometheusExtConf struct {
	Host string
	Port int
	Path string
}
