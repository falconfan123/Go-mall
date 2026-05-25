package config

import (
	"github.com/falconfan123/Go-mall/common/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	// gRPC 配置
	zrpc.RpcServerConf
	PostgresConfig PostgresConfig
	RedisConf      redis.RedisConf
	ElasticSearch  config.ElasticSearchConfig
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

type PrometheusExtConf struct {
	Host string
	Port int
	Path string
}
