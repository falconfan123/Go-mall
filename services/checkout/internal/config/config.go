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
	CheckoutRpc    zrpc.RpcClientConf
	CouponRpc      zrpc.RpcClientConf
	UserRpc        zrpc.RpcClientConf
	InventoryRpc   zrpc.RpcClientConf
	ProductRpc     zrpc.RpcClientConf
	CouponsRpc     zrpc.RpcClientConf
	RabbitMQConfig commonconfig.RabbitMQConfig
}

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}
