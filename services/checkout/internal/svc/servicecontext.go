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
	Mysql              sqlx.SqlConn
	RedisClient        *redis.Redis
	CheckoutModel      checkout.CheckoutsModel
	CheckoutItemsModel checkout.CheckoutItemsModel
	CartsModel         cart.CartsModel
	InventoryRpc       inventoryclient.Inventory
	CouponsRpc         couponsclient.Coupons
	ProductRpc         productclient.ProductCatalog
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
		InventoryRpc:       inventory.NewInventoryClient(zrpc.MustNewClient(c.InventoryRpc)),
		CouponsRpc:         coupons.NewCouponsClient(zrpc.MustNewClient(c.CouponsRpc)),
		ProductRpc:         product.NewProductCatalogClient(zrpc.MustNewClient(c.ProductRpc)),
	}
}
