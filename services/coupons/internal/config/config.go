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
}

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}
