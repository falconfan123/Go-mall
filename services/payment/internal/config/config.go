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
	PostgresConfig config.PostgresConfig
	RedisConf      redis.RedisConf
	Alipay         AlipayConfig
	Stripe         StripeConfig
	OrderRpc       zrpc.RpcClientConf
	RabbitMQConfig config.RabbitMQConfig
}

type AlipayConfig struct {
	AppId           string
	PrivateKey      string
	AlipayPublicKey string
	NotifyURL       string
	NotifyPath      string
	NotifyPort      int
	ReturnURL       string
}

type StripeConfig struct {
	APIKey        string
	SuccessURL    string
	CancelURL     string
	WebhookSecret string
}
