package config

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	RedisConf      redis.RedisConf

	InventoryRpc   zrpc.RpcClientConf
	CouponsRpc     zrpc.RpcClientConf
	ProductRpc     zrpc.RpcClientConf
	CheckoutRpc    zrpc.RpcClientConf
	CouponRpc      zrpc.RpcClientConf
	UserRpc        zrpc.RpcClientConf
	RabbitMQConfig RabbitMQConfig
	PrometheusExt  PrometheusExtConf
}

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}

type RabbitMQConfig struct {
	Host  string
	Port  int
	User  string
	Pass  string
	VHost string
}

func (r *RabbitMQConfig) Dns() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", r.User, r.Pass, r.Host, r.Port, r.VHost)
}

type PrometheusExtConf struct {
	Host string
	Port int
	Path string
}
