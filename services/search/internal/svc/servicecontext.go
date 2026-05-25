package svc

import (
	"database/sql"

	"github.com/falconfan123/Go-mall/services/search/internal/config"
	"github.com/falconfan123/Go-mall/services/search/internal/db"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/service"
	"github.com/falconfan123/Go-mall/services/search/internal/infrastructure/persistence"
	"github.com/falconfan123/Go-mall/services/search/internal/stats"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config              config.Config
	DB                  *sql.DB
	QueryParser         *service.QueryParser
	ElasticsearchClient *persistence.ElasticsearchClient
	RAGStats            *stats.Service
	StatsHTTPServer     *stats.HTTPServer
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 Elasticsearch 客户端
	esClient := persistence.NewElasticsearchClient(c.ElasticSearch)

	// 初始化 QueryParser
	queryParser := service.NewQueryParser(c.RedisConf)
	postgresDB := db.NewPostgres(c.PostgresConfig)
	statsStore := stats.NewPostgresStore(postgresDB)
	statsMetrics := stats.NewMetrics(nil)
	statsService := stats.NewService(statsStore, statsMetrics)
	statsHTTPServer := stats.NewHTTPServer(c.PrometheusExt, statsService)

	return &ServiceContext{
		Config:              c,
		DB:                  postgresDB,
		QueryParser:         queryParser,
		ElasticsearchClient: esClient,
		RAGStats:            statsService,
		StatsHTTPServer:     statsHTTPServer,
	}
}

func MustNewRedis(c config.Config) redis.RedisConf {
	return c.RedisConf
}
