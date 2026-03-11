package svc

import (
	"github.com/falconfan123/Go-mall/dal/model/cart"
	"github.com/falconfan123/Go-mall/dal/model/checkout"
	"github.com/falconfan123/Go-mall/services/checkout/internal/config"
	"github.com/falconfan123/Go-mall/services/checkout/internal/db"
	"github.com/falconfan123/Go-mall/services/coupons/couponsclient"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/product/product"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config             config.Config
	Mysql              sqlx.SqlConn
	RedisClient        *redis.Redis
	CheckoutModel      checkout.CheckoutsModel
	CheckoutItemsModel checkout.CheckoutItemsModel
	CartsModel         cart.CartsModel
	InventoryRpc       inventoryclient.Inventory
	CouponsRpc         couponsclient.Coupons
	ProductRpc         product.ProductCatalogService
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysql := db.NewMysql(c.MysqlConfig)
	redisconf, err := redis.NewRedis(c.RedisConf)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:             c,
		Mysql:              mysql,
		RedisClient:        redisconf,
		CartsModel:         cart.NewCartsModel(mysql),
		CheckoutModel:      checkout.NewCheckoutsModel(mysql),
		CheckoutItemsModel: checkout.NewCheckoutItemsModel(mysql),
		InventoryRpc:       inventoryclient.NewInventory(zrpc.MustNewClient(c.InventoryRpc)),
		CouponsRpc:         couponsclient.NewCoupons(zrpc.MustNewClient(c.CouponsRpc)),
		ProductRpc:         product.NewProductCatalogService(zrpc.MustNewClient(c.ProductRpc)),
	}
}
