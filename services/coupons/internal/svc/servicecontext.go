package svc

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/falconfan123/Go-mall/dal/model/coupons/coupon"
	"github.com/falconfan123/Go-mall/dal/model/coupons/coupon_usage"
	"github.com/falconfan123/Go-mall/dal/model/coupons/user_coupons"
	"github.com/falconfan123/Go-mall/services/coupons/internal/config"
	productclient "github.com/falconfan123/Go-mall/services/product/productclient"
	"github.com/zeromicro/go-zero/core/logx"
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
	// DtmDB barrier 专用连接（dtm BranchBarrier.Call 自管事务；与业务共用同一 DSN）
	DtmDB      *sql.DB
	Rdb        *redis.Redis
	ProductRpc productclient.ProductCatalog
}

func NewServiceContext(c config.Config) *ServiceContext {
	dtmDB, err := sql.Open("postgres", c.PostgresConfig.DataSource)
	if err != nil {
		logx.Errorf("open dtm barrier db failed: %v", err)
		panic(err)
	}
	dtmDB.SetMaxOpenConns(20)

	return &ServiceContext{
		Config:           c,
		DtmDB:            dtmDB,
		CouponsModel:     coupon.NewCouponsModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		UserCouponsModel: user_coupons.NewUserCouponsModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		CouponUsageModel: coupon_usage.NewCouponUsageModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		Model:            sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource),
		Rdb:              redis.MustNewRedis(c.RedisConf),
		ProductRpc:       productclient.NewProductCatalog(zrpc.MustNewClient(c.ProductRpc)),
	}
}
