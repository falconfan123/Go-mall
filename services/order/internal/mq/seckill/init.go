package seckill

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/order/internal/config"
	"github.com/streadway/amqp"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	SeckillExchangeName = "seckill-order-exchange"
	SeckillQueueName    = "seckill-order-queue"
)

type SeckillOrder struct {
	OrderID    string `json:"order_id"`
	UserID     int64  `json:"user_id"`
	ProductID  int64  `json:"product_id"`
	ActivityID int64  `json:"activity_id"`
	Timestamp  int64  `json:"timestamp"`
}

type SeckillMQ struct {
	conn           *amqp.Connection
	OrderModel     order.OrdersModel
	OrderItemModel order.OrderItemsModel
	Model          sqlx.SqlConn
}

func Init(c config.Config) (*SeckillMQ, error) {
	conn, err := amqp.Dial(c.RabbitMQConfig.Dns())
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// 声明交换机
	err = ch.ExchangeDeclare(
		SeckillExchangeName,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// 声明队列
	_, err = ch.QueueDeclare(
		SeckillQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// 绑定队列到交换机
	err = ch.QueueBind(
		SeckillQueueName,
		"",
		SeckillExchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	mq := &SeckillMQ{
		conn:           conn,
		OrderModel:     order.NewOrdersModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		OrderItemModel: order.NewOrderItemsModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		Model:          sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource),
	}

	// 启动消费者
	go mq.consumer(context.Background())

	return mq, nil
}

// Publish 发布秒杀订单消息
func (m *SeckillMQ) Publish(msg SeckillOrder) error {
	ch, err := m.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return ch.Publish(
		SeckillExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

// consumer 消费秒杀订单消息
func (m *SeckillMQ) consumer(ctx context.Context) {
	ch, err := m.conn.Channel()
	if err != nil {
		log.Printf("seckill mq consumer channel error: %v", err)
		return
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		SeckillQueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("seckill mq consume error: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}

			var seckillOrder SeckillOrder
			if err := json.Unmarshal(d.Body, &seckillOrder); err != nil {
				log.Printf("seckill mq unmarshal error: %v", err)
				d.Nack(false, false)
				continue
			}

			// 创建订单
			if err := m.createOrder(seckillOrder); err != nil {
				log.Printf("seckill order create error: %v", err)
				d.Nack(false, true) // 重新入队
				continue
			}

			d.Ack(false)
		}
	}
}

// createOrder 创建订单
func (m *SeckillMQ) createOrder(seckillOrder SeckillOrder) error {
	// 生成订单号
	orderID := seckillOrder.OrderID
	if orderID == "" {
		orderID = fmt.Sprintf("SK%d%d%d", time.Now().UnixMilli(), seckillOrder.UserID, seckillOrder.ProductID)
	}

	// 创建订单记录
	orderRecord := &order.Orders{
		OrderId:        orderID,
		PreOrderId:     orderID,
		UserId:         uint64(seckillOrder.UserID),
		OrderStatus:    1, // 创建（待支付）
		PaymentStatus:  1, // 未支付
		PayableAmount:  1, // 秒杀价 1 分
		OriginalAmount: 1, // 秒杀价 1 分
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := m.OrderModel.Insert(context.Background(), orderRecord)
	if err != nil {
		return fmt.Errorf("insert order error: %v", err)
	}

	// 创建订单项
	orderItem := &order.OrderItems{
		OrderId:     orderID,
		ProductId:   uint64(seckillOrder.ProductID),
		ProductName: fmt.Sprintf("秒杀商品-%d", seckillOrder.ProductID),
		Quantity:    1,
		CreatedAt:   time.Now(),
	}

	_, err = m.OrderItemModel.Insert(context.Background(), orderItem)
	if err != nil {
		return fmt.Errorf("insert order item error: %v", err)
	}

	return nil
}
