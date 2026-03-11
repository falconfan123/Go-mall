package svc

import (
	"github.com/falconfan123/Go-mall/dal/model/coupons/coupon"
	"github.com/falconfan123/Go-mall/dal/model/coupons/coupon_usage"
	"github.com/falconfan123/Go-mall/dal/model/coupons/user_coupons"
	"github.com/falconfan123/Go-mall/services/coupons/internal/config"
	"github.com/falconfan123/Go-mall/services/product/product"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config           config.Config
	CouponsModel     coupon.CouponsModel
	UserCouponsModel user_coupons.UserCouponsModel
	CouponUsageModel coupon_usage.CouponUsageModel
	Model            sqlx.SqlConn
	Rdb              *redis.Redis
	ProductRpc       product.ProductCatalogService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:           c,
		CouponsModel:     coupon.NewCouponsModel(sqlx.NewMysql(c.MysqlConfig.DataSource)),
		UserCouponsModel: user_coupons.NewUserCouponsModel(sqlx.NewMysql(c.MysqlConfig.DataSource)),
		CouponUsageModel: coupon_usage.NewCouponUsageModel(sqlx.NewMysql(c.MysqlConfig.DataSource)),
		Model:            sqlx.NewMysql(c.MysqlConfig.DataSource),
		Rdb:              redis.MustNewRedis(c.RedisConf),
		ProductRpc:       product.NewProductCatalogService(zrpc.MustNewClient(c.ProductRpc)),
	}
}
