package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf
	Consul     consul.Conf
	RedisConf  redis.RedisConf
	Activity   ActivityConfig
	Prometheus PrometheusConf
}

type ActivityConfig struct {
	TokenExpire    int // token 有效期（秒）
	AdvanceSeconds int // 提前获取 token 的秒数
}

type PrometheusConf struct {
	Host string
	Port int
	Path string
}
