package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	RedisConf redis.RedisConf
	Activity  ActivityConfig
}

type ActivityConfig struct {
	TokenExpire    int // token 有效期（秒）
	AdvanceSeconds int // 提前获取 token 的秒数
}
