package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	GorseConfig    GorseConfig
	AuditRpc       zrpc.RpcClientConf
	AuthsRpc       zrpc.RpcClientConf
	Consul         consul.Conf
	Cache          cache.CacheConf
	RedisConf      redis.RedisConf
	Salt           string
	AuthConfig     struct {
		AccessSecret string
		AccessExpire int64
	}
}
type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}
type GorseConfig struct {
	GorseAddr   string
	GorseApikey string
}
