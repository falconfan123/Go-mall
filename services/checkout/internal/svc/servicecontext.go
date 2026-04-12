package svc

import (
	"github.com/falconfan123/Go-mall/dal/model/cart"
	"github.com/falconfan123/Go-mall/dal/model/checkout"
	"github.com/falconfan123/Go-mall/services/checkout/internal/config"
	"github.com/falconfan123/Go-mall/services/checkout/internal/db"
	couponsclient "github.com/falconfan123/Go-mall/services/coupons/couponsclient"
	inventoryclient "github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	productclient "github.com/falconfan123/Go-mall/services/product/productclient"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config             config.Config
	Postgres           sqlx.SqlConn
	RedisClient        *redis.Redis
	CheckoutModel      checkout.CheckoutsModel
	CheckoutItemsModel checkout.CheckoutItemsModel
	CartsModel         cart.CartsModel
	InventoryRpc       inventoryclient.Inventory
	CouponsRpc         couponsclient.Coupons
	ProductRpc         productclient.ProductCatalog
}

func NewServiceContext(c config.Config) *ServiceContext {
	postgres := db.NewPostgres(c.PostgresConfig)
	redisconf, err := redis.NewRedis(c.RedisConf)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:             c,
		Postgres:           postgres,
		RedisClient:        redisconf,
		CartsModel:         cart.NewCartsModel(postgres),
		CheckoutModel:      checkout.NewCheckoutsModel(postgres),
		CheckoutItemsModel: checkout.NewCheckoutItemsModel(postgres),
		InventoryRpc:       inventoryclient.NewInventory(zrpc.MustNewClient(c.InventoryRpc)),
		CouponsRpc:         couponsclient.NewCoupons(zrpc.MustNewClient(c.CouponsRpc)),
		ProductRpc:         productclient.NewProductCatalog(zrpc.MustNewClient(c.ProductRpc)),
	}
}
