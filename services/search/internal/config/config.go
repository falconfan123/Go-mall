package config

import (
	"github.com/falconfan123/Go-mall/common/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	// gRPC 配置
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	RedisConf      redis.RedisConf
	ElasticSearch  config.ElasticSearchConfig
	ProductRpc     zrpc.RpcClientConf
}

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}
