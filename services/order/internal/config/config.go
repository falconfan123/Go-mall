package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	RedisConf      redis.RedisConf
	Consul         consul.Conf

	InventoryRpc   zrpc.RpcClientConf
	CouponsRpc     zrpc.RpcClientConf
	ProductRpc     zrpc.RpcClientConf
	CheckoutRpc    zrpc.RpcClientConf
	CouponRpc      zrpc.RpcClientConf
	UserRpc        zrpc.RpcClientConf
	RabbitMQConfig RabbitMQConfig
	Prometheus     PrometheusConf
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
	return "amqp://" + r.User + ":" + r.Pass + "@" + r.Host + ":" + string(rune(r.Port)) + "/" + r.VHost
}

type PrometheusConf struct {
	Host string
	Port int
	Path string
}
