package svc

import (
	"context"

	_ "github.com/lib/pq"

	"github.com/falconfan123/Go-mall/dal/model/payment"
	order "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/falconfan123/Go-mall/services/payment/internal/mq"
	"github.com/falconfan123/Go-mall/services/payment/internal/stripe"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// TxRunner 事务执行抽象（webhook 两写事务可注入 fake，Rule 6）。
type TxRunner interface {
	TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error
}

type ServiceContext struct {
	Config          config.Config
	Rdb             *redis.Redis
	PaymentModel    payment.PaymentsModel
	OutboxModel     payment.PaymentOutboxModel
	ExceptionModel  payment.SettlementExceptionModel
	OrderRpc        order.OrderServiceClient
	StripeProcessor *stripe.StripeProcessor
	PaymentMQ       *mq.PaymentDelayMQ
	Model           TxRunner
	// RawConn 只读连接（扫描器跨域读，对账作业惯例；写路径全部走服务 RPC）
	RawConn sqlx.SqlConn
	// SettlementSaga 结算 saga 发起/幂等补建（design 决策 1）
	SettlementSaga SettlementSagaAPI
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
		OutboxModel:     payment.NewPaymentOutboxModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		ExceptionModel:  payment.NewSettlementExceptionModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		OrderRpc:        order.NewOrderServiceClient(zrpc.MustNewClient(c.OrderRpc).Conn()),
		StripeProcessor: stripeProcessor,
		PaymentMQ:       nil, // 暂时设置为nil（决策 6：保持休眠）
		Model:           sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource),
		RawConn:         sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource),
		SettlementSaga:  NewSettlementSaga(c.Dtm, logx.WithContext(context.Background())),
	}
}
