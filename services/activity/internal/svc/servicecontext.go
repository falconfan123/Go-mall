package svc

import (
	"github.com/falconfan123/Go-mall/services/activity/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	r, _ := redis.NewRedis(c.RedisConf)
	return &ServiceContext{
		Config: c,
		Redis:  r,
	}
}
