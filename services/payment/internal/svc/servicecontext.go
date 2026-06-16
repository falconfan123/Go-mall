package svc

import (
	_ "github.com/lib/pq"

	"github.com/falconfan123/Go-mall/dal/model/payment"
	order "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/falconfan123/Go-mall/services/payment/internal/mq"
	"github.com/falconfan123/Go-mall/services/payment/internal/stripe"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config          config.Config
	Rdb             *redis.Redis
	PaymentModel    payment.PaymentsModel
	OrderRpc        order.OrderServiceClient
	StripeProcessor *stripe.StripeProcessor
	PaymentMQ       *mq.PaymentDelayMQ
	Model           sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 暂时注释延迟队列初始化，因为RabbitMQ插件未安装
	// delayMQ, err := mq.Init(c)
	// if err != nil {
	// 	logx.Errorw("创建延迟队列失败", logx.LogField{Key: "err", Value: err})
	// 	panic(err)
	// }
	stripeProcessor := stripe.NewStripeProcessor(c.Stripe)

	return &ServiceContext{
		Config:          c,
		Rdb:             redis.MustNewRedis(c.RedisConf),
		PaymentModel:    payment.NewPaymentsModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		OrderRpc:        order.NewOrderServiceClient(zrpc.MustNewClient(c.OrderRpc).Conn()),
		StripeProcessor: stripeProcessor,
		PaymentMQ:       nil, // 暂时设置为nil
		Model:           sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource),
	}
}
