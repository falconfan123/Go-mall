package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/falconfan123/Go-mall/dal/model/cart"
	"github.com/falconfan123/Go-mall/services/carts/internal/config"
	"github.com/falconfan123/Go-mall/services/carts/internal/db"
)

type ServiceContext struct {
	Config     config.Config
	Mysql      sqlx.SqlConn
	CartsModel cart.CartsModel
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	mysql := db.NewMysql(c.MysqlConfig)
	return &ServiceContext{
		Config:     c,
		Mysql:      mysql,
		CartsModel: cart.NewCartsModel(mysql),
	}, nil
}
