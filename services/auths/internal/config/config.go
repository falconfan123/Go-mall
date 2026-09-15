package config

import (
	"github.com/falconfan123/Go-mall/common/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig config.PostgresConfig
	SessionRedis   redis.RedisConf
}
