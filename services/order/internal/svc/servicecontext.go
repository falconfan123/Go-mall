package svc

import (
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
	orderDelayMQ, err := delay.Init(c)
	if err != nil {
		logx.Errorf("delay mq init failed: %v, continuing without it", err)
		orderDelayMQ = nil
	}
	notifyMQ, err := notify.Init(c)
	if err != nil {
		logx.Errorf("notify mq init failed: %v, continuing without it", err)
		notifyMQ = nil
	}
	seckillMQ, err := seckill.Init(c)
	if err != nil {
		logx.Errorf("seckill mq init failed: %v, continuing without it", err)
		seckillMQ = nil
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
