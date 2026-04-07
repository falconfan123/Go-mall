package config

import (
	"github.com/falconfan123/Go-mall/common/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	// gRPC 配置
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	RedisConf      redis.RedisConf
	ElasticSearch  config.ElasticSearchConfig
	QiNiu          QiNiu
	Consul         consul.Conf
	InventoryRpc   zrpc.RpcClientConf
	GorseConfig    config.GorseConfig
	Minio          Minio
	PrometheusExt  PrometheusExtConf
}

type Minio struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}

type QiNiu struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

type PrometheusExtConf struct {
	Host string
	Port int
	Path string
}
