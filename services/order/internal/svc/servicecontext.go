package svc

import (
	"time"

	"github.com/avast/retry-go"
	"github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/checkout/checkoutservice"
	"github.com/falconfan123/Go-mall/services/coupons/couponsclient"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/order/internal/config"
	"github.com/falconfan123/Go-mall/services/order/internal/mq/delay"
	"github.com/falconfan123/Go-mall/services/order/internal/mq/notify"
	"github.com/falconfan123/Go-mall/services/order/internal/mq/seckill"
	userspb "github.com/falconfan123/Go-mall/services/users/pb"
	"github.com/zeromicro/go-zero/core/logx"
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
	OrderDelayMQ   *delay.OrderDelayMQ
	OrderNotifyMQ  *notify.OrderNotifyMQ
	SeckillMQ      *seckill.SeckillMQ
	RedisClient    *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	orderDelayMQ, err := initOrderDelayMQ(c)
	if err != nil {
		logx.Errorf("delay mq init failed after retries: %v", err)
		panic(err)
	}
	notifyMQ, err := initOrderNotifyMQ(c)
	if err != nil {
		logx.Errorf("notify mq init failed after retries: %v", err)
		panic(err)
	}
	seckillMQ, err := initSeckillMQ(c)
	if err != nil {
		logx.Errorf("seckill mq init failed after retries: %v", err)
		panic(err)
	}
	redisClient, err := redis.NewRedis(c.RedisConf)
	if err != nil {
		logx.Error(err)
		panic(err)
	}
	return &ServiceContext{
		Config:         c,
		OrderModel:     order.NewOrdersModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		OrderItemModel: order.NewOrderItemsModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		OrderAddress:   order.NewOrderAddressesModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		Model:          sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource),
		CheckoutRpc:    checkoutservice.NewCheckoutService(zrpc.MustNewClient(c.CheckoutRpc)),
		CouponRpc:      couponsclient.NewCoupons(zrpc.MustNewClient(c.CouponRpc)),
		UserRpc:        userspb.NewUsersClient(zrpc.MustNewClient(c.UserRpc).Conn()),
		InventoryRpc:   inventoryclient.NewInventory(zrpc.MustNewClient(c.InventoryRpc)),
		OrderDelayMQ:   orderDelayMQ,
		OrderNotifyMQ:  notifyMQ,
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

func initOrderNotifyMQ(c config.Config) (*notify.OrderNotifyMQ, error) {
	var (
		orderNotifyMQ *notify.OrderNotifyMQ
		err           error
	)

	retryErr := retry.Do(
		func() error {
			orderNotifyMQ, err = notify.Init(c)
			return err
		},
		retry.Attempts(30),
		retry.Delay(2*time.Second),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			logx.Errorf("notify mq init attempt %d/30 failed: %v", n+1, err)
		}),
	)
	if retryErr != nil {
		return nil, retryErr
	}

	return orderNotifyMQ, nil
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
