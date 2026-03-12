package delay

import (
	"context"
	"github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/checkout/checkoutservice"
	"github.com/falconfan123/Go-mall/services/coupons/couponsclient"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/order/internal/config"
	"github.com/streadway/amqp"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"time"
)

const (
	ExchangeName       = "order-delay-exchange"
	ExchangeKind       = amqp.ExchangeDirect
	QueueName          = "order-delay-queue"
	DeadLetterExchange = "order-delay-dlx-exchange"
	DeadLetterQueue    = "order-delay-dlx-queue"
	Delay              = 30 * time.Minute
)

type OrderDelayMQ struct {
	conn            *amqp.Connection
	OrderModel      order.OrdersModel
	OrderItemsModel order.OrderItemsModel
	Model           sqlx.SqlConn
	CheckoutRpc     checkoutservice.CheckoutService
	CouponRpc       couponsclient.Coupons
	InventoryRpc    inventoryclient.Inventory
}
type OrderReq struct {
	OrderId  string `json:"order_id"`
	UserID   int32  `json:"user_id"`
	RetryCnt int
}

func Init(c config.Config) (*OrderDelayMQ, error) {
	conn, err := amqp.Dial(c.RabbitMQConfig.Dns())
	if err != nil {
		return nil, err
	}
	// 创建通道
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	// 声明交换机
	err = ch.ExchangeDeclare(
		ExchangeName, // 交换机名称
		ExchangeKind, // 类型为直接交换机
		true,         // 持久化
		false,        // 自动删除
		false,        // 内部交换机
		false,        // 等待确认
		nil,
	)
	if err != nil {
		return nil, err
	}
	// 声明死信交换机
	err = ch.ExchangeDeclare(
		DeadLetterExchange,  // 死信交换机名称
		amqp.ExchangeDirect, // 类型为直接交换机
		true,                // 持久化
		false,               // 自动删除
		false,               // 内部交换机
		false,               // 等待确认
		nil,
	)
	if err != nil {
		return nil, err
	}
	// 声明死信队列
	_, err = ch.QueueDeclare(
		DeadLetterQueue, // 死信队列名称
		true,            // 持久化
		false,           // 自动删除
		false,           // 排他性
		false,           // 等待确认
		nil,
	)
	if err != nil {
		return nil, err
	}
	// 绑定死信队列到死信交换机
	err = ch.QueueBind(
		DeadLetterQueue,
		"",
		DeadLetterExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	// 声明延迟队列（包含死信交换机参数）
	_, err = ch.QueueDeclare(
		QueueName, // 队列名称
		true,      // 持久化
		false,     // 自动删除
		false,     // 排他性
		false,     // 等待确认
		amqp.Table{
			"x-dead-letter-exchange":    DeadLetterExchange,
			"x-dead-letter-routing-key": "",
		},
	)
	if err != nil {
		return nil, err
	}
	// 绑定队列到交换机
	err = ch.QueueBind(
		QueueName,
		"",
		ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	orderDelay := &OrderDelayMQ{
		conn:            conn,
		OrderModel:      order.NewOrdersModel(sqlx.NewMysql(c.MysqlConfig.DataSource)),
		CheckoutRpc:     checkoutservice.NewCheckoutService(zrpc.MustNewClient(c.CheckoutRpc)),
		CouponRpc:       couponsclient.NewCoupons(zrpc.MustNewClient(c.CouponRpc)),
		InventoryRpc:    inventoryclient.NewInventory(zrpc.MustNewClient(c.InventoryRpc)),
		Model:           sqlx.NewMysql(c.MysqlConfig.DataSource),
		OrderItemsModel: order.NewOrderItemsModel(sqlx.NewMysql(c.MysqlConfig.DataSource)),
	}
	go orderDelay.consumer(context.TODO())
	return orderDelay, nil
}
