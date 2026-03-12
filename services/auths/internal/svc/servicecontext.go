package svc

import (
	"github.com/falconfan123/Go-mall/dal/model/user"
	"github.com/falconfan123/Go-mall/services/auths/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel user.UsersModel
	Redis     *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)
	return &ServiceContext{
		UserModel: user.NewUsersModel(conn),
		Config:    c,
		Redis:     redis.MustNewRedis(c.SessionRedis),
	}
}
