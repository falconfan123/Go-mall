package svc

import (
	"github.com/falconfan123/Go-mall/services/search/internal/config"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/service"
	"github.com/falconfan123/Go-mall/services/search/internal/infrastructure/persistence"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config              config.Config
	QueryParser         *service.QueryParser
	ElasticsearchClient *persistence.ElasticsearchClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 Elasticsearch 客户端
	esClient := persistence.NewElasticsearchClient(c.ElasticSearch)

	// 初始化 QueryParser
	queryParser := service.NewQueryParser(c.RedisConf)

	return &ServiceContext{
		Config:              c,
		QueryParser:         queryParser,
		ElasticsearchClient: esClient,
	}
}

func MustNewRedis(c config.Config) redis.RedisConf {
	return c.RedisConf
}
