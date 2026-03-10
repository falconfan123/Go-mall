package svc

import (
	"github.com/falconfan123/Go-mall/dal/model/user"
	"github.com/falconfan123/Go-mall/services/users/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	UsersModel user.UsersModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:     c,
		UsersModel: user.NewUsersModel(conn),
	}
}
