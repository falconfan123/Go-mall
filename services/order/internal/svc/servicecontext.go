package svc

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"github.com/avast/retry-go"
	"github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/checkout/checkoutservice"
	"github.com/falconfan123/Go-mall/services/coupons/couponsclient"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/order/internal/config"
	"github.com/falconfan123/Go-mall/services/order/internal/mq/delay"
	"github.com/falconfan123/Go-mall/services/order/internal/mq/seckill"
	userspb "github.com/falconfan123/Go-mall/services/users/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	OrderModel     order.OrdersModel
	OrderItemModel order.OrderItemsModel
	OrderAddress   order.OrderAddressesModel
	CheckoutRpc    checkoutservice.CheckoutService
	CouponRpc      couponsclient.Coupons
	UserRpc        userspb.UsersClient
	InventoryRpc   inventoryclient.Inventory
	Model          sqlx.SqlConn
	// DtmDB barrier 专用连接（dtm BranchBarrier.Call 自管事务；与业务共用同一 DSN）
	DtmDB        *sql.DB
	OrderDelayMQ *delay.OrderDelayMQ
	SeckillMQ    *seckill.SeckillMQ
	RedisClient  *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 结算 consumer 生命周期：ctx 由服务持有，wrap-up 阶段 cancel 触发优雅退出
	consumerCtx, cancelConsumers := context.WithCancel(context.Background())
	proc.AddWrapUpListener(cancelConsumers)

	orderDelayMQ, err := initOrderDelayMQ(c)
	if err != nil {
		logx.Errorf("delay mq init failed after retries: %v", err)
		panic(err)
	}
	seckillMQ, err := initSeckillMQ(c)
	if err != nil {
		logx.Errorf("seckill mq init failed after retries: %v", err)
		panic(err)
	}
	// 结算 consumer 接入监督骨架（autoAck=false + 优雅停机）
	orderDelayMQ.Start(consumerCtx)
	redisClient, err := redis.NewRedis(c.RedisConf)
	if err != nil {
		logx.Error(err)
		panic(err)
	}
	dtmDB, err := sql.Open("postgres", c.PostgresConfig.DataSource)
	if err != nil {
		logx.Errorf("open dtm barrier db failed: %v", err)
		panic(err)
	}
	dtmDB.SetMaxOpenConns(20)
	return &ServiceContext{
		Config:         c,
		DtmDB:          dtmDB,
		OrderModel:     order.NewOrdersModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		OrderItemModel: order.NewOrderItemsModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		OrderAddress:   order.NewOrderAddressesModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		Model:          sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource),
		CheckoutRpc:    checkoutservice.NewCheckoutService(zrpc.MustNewClient(c.CheckoutRpc)),
		CouponRpc:      couponsclient.NewCoupons(zrpc.MustNewClient(c.CouponRpc)),
		UserRpc:        userspb.NewUsersClient(zrpc.MustNewClient(c.UserRpc).Conn()),
		InventoryRpc:   inventoryclient.NewInventory(zrpc.MustNewClient(c.InventoryRpc)),
		OrderDelayMQ:   orderDelayMQ,
		SeckillMQ:      seckillMQ,
		RedisClient:    redisClient,
	}
}

func initOrderDelayMQ(c config.Config) (*delay.OrderDelayMQ, error) {
	var (
		orderDelayMQ *delay.OrderDelayMQ
		err          error
	)

	retryErr := retry.Do(
		func() error {
			orderDelayMQ, err = delay.Init(c)
			return err
		},
		retry.Attempts(30),
		retry.Delay(2*time.Second),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			logx.Errorf("delay mq init attempt %d/30 failed: %v", n+1, err)
		}),
	)
	if retryErr != nil {
		return nil, retryErr
	}

	return orderDelayMQ, nil
}

func initSeckillMQ(c config.Config) (*seckill.SeckillMQ, error) {
	var (
		seckillMQ *seckill.SeckillMQ
		err       error
	)

	retryErr := retry.Do(
		func() error {
			seckillMQ, err = seckill.Init(c)
			return err
		},
		retry.Attempts(30),
		retry.Delay(2*time.Second),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			logx.Errorf("seckill mq init attempt %d/30 failed: %v", n+1, err)
		}),
	)
	if retryErr != nil {
		return nil, retryErr
	}

	return seckillMQ, nil
}
