package config

import (
	"github.com/falconfan123/Go-mall/common/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig config.PostgresConfig
	RedisConf      redis.RedisConf
	ProductRpc     zrpc.RpcClientConf
	InventoryRpc   zrpc.RpcClientConf
	PrometheusExt  PrometheusExtConf
}

type PrometheusExtConf struct {
	Host string
	Port int
	Path string
}
