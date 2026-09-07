package config

import (
	commonconfig "github.com/falconfan123/Go-mall/common/config"

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
	RabbitMQConfig commonconfig.RabbitMQConfig
	Consumer       commonconfig.ConsumerConfig
}

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}
