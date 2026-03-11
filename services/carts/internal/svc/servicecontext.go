package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/falconfan123/Go-mall/dal/model/cart"
	"github.com/falconfan123/Go-mall/services/carts/internal/application/service"
	"github.com/falconfan123/Go-mall/services/carts/internal/config"
	"github.com/falconfan123/Go-mall/services/carts/internal/db"
	"github.com/falconfan123/Go-mall/services/carts/internal/infrastructure/persistence"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/repository"
)

type ServiceContext struct {
	Config         config.Config
	Mysql          sqlx.SqlConn
	CartsModel     cart.CartsModel
	CartRepo       repository.CartRepository
	CartAppService *service.CartAppService
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	mysql := db.NewMysql(c.MysqlConfig)
	cartsModel := cart.NewCartsModel(mysql)
	cartRepo := persistence.NewCartRepositoryImpl(cartsModel)
	cartAppService := service.NewCartAppService(cartRepo)

	return &ServiceContext{
		Config:         c,
		Mysql:          mysql,
		CartsModel:     cartsModel,
		CartRepo:       cartRepo,
		CartAppService: cartAppService,
	}, nil
}
