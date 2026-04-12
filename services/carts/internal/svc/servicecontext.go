package svc

import (
	"github.com/falconfan123/Go-mall/dal/model/cart"
	"github.com/falconfan123/Go-mall/services/carts/internal/application/service"
	"github.com/falconfan123/Go-mall/services/carts/internal/config"
	"github.com/falconfan123/Go-mall/services/carts/internal/db"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/carts/internal/infrastructure/persistence"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	Postgres       sqlx.SqlConn
	CartsModel     cart.CartsModel
	CartRepo       repository.CartRepository
	CartAppService *service.CartAppService
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	postgres := db.NewPostgres(c.PostgresConfig)
	cartsModel := cart.NewCartsModel(postgres)
	cartRepo := persistence.NewCartRepositoryImpl(cartsModel)
	cartAppService := service.NewCartAppService(cartRepo)

	return &ServiceContext{
		Config:         c,
		Postgres:       postgres,
		CartsModel:     cartsModel,
		CartRepo:       cartRepo,
		CartAppService: cartAppService,
	}, nil
}
