package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
	"github.com/falconfan123/Go-mall/common/config"
)

type Config struct {
	zrpc.RpcServerConf
	MysqlConfig    config.MysqlConfig
	RedisConf      redis.RedisConf
	Consul         consul.Conf
	CheckoutRpc    zrpc.RpcClientConf
	CouponRpc      zrpc.RpcClientConf
	UserRpc        zrpc.RpcClientConf
	InventoryRpc   zrpc.RpcClientConf
	RabbitMQConfig config.RabbitMQConfig
}
