package config

import (
	"github.com/falconfan123/Go-mall/common/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf
	Consul         consul.Conf
	RabbitMQ       config.RabbitMQConfig
	PostgresConfig config.PostgresConfig
	ElasticSearch  config.ElasticSearchConfig
	RedisConf      redis.RedisConf
	PrometheusExt  PrometheusExtConf
}

type PrometheusExtConf struct {
	Host string
	Port int
	Path string
}
