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
	Stripe         StripeConfig
	OrderRpc       zrpc.RpcClientConf
	RabbitMQConfig config.RabbitMQConfig
	PrometheusExt  PrometheusExtConf
}

type StripeConfig struct {
	APIKey            string
	SuccessURL        string
	CancelURL         string
	WebhookSecret     string
	WebhookPort       int
	RequestTimeoutMs  int64
	MaxNetworkRetries int64
}

type PrometheusExtConf struct {
	Host string
	Port int
	Path string
}
