package messaging

import (
	"encoding/json"
	"fmt"

	"github.com/falconfan123/Go-mall/services/product/internal/application/event"
	domainevent "github.com/falconfan123/Go-mall/services/product/internal/domain/event"
	"github.com/streadway/amqp"
)

// RabbitMQEventPublisher RabbitMQ事件发布器实现
type RabbitMQEventPublisher struct {
	channel *amqp.Channel
}

// NewRabbitMQEventPublisher 创建RabbitMQ事件发布器
func NewRabbitMQEventPublisher(channel *amqp.Channel) event.ProductEventPublisher {
	// 声明交换机
	err := channel.ExchangeDeclare(
		"product_events", // 交换机名称
		"topic",          // 交换机类型
		true,             // 持久化
		false,            // 自动删除
		false,            // 内部使用
		false,            // 不等待
		nil,              // 其他参数
	)
	if err != nil {
		fmt.Printf("failed to declare exchange: %v\n", err)
	}

	return &RabbitMQEventPublisher{
		channel: channel,
	}
}

// PublishProductCreated 发布商品创建事件
func (p *RabbitMQEventPublisher) PublishProductCreated(evt *domainevent.ProductCreatedEvent) error {
	return p.publishEvent("product.created", evt)
}

// PublishProductUpdated 发布商品更新事件
func (p *RabbitMQEventPublisher) PublishProductUpdated(evt *domainevent.ProductUpdatedEvent) error {
	return p.publishEvent("product.updated", evt)
}

// PublishProductStockChanged 发布商品库存变化事件
func (p *RabbitMQEventPublisher) PublishProductStockChanged(evt *domainevent.ProductStockChangedEvent) error {
	return p.publishEvent("product.stock_changed", evt)
}

// PublishProductDeleted 发布商品删除事件
func (p *RabbitMQEventPublisher) PublishProductDeleted(evt *domainevent.ProductDeletedEvent) error {
	return p.publishEvent("product.deleted", evt)
}

// publishEvent 发布事件通用方法
func (p *RabbitMQEventPublisher) publishEvent(routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.channel.Publish(
		"product_events", // 交换机名称
		routingKey,       // 路由键
		false,            // 强制
		false,            // 立即
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		fmt.Printf("failed to publish event %s: %v\n", routingKey, err)
		return err
	}

	fmt.Printf("published event %s: %s\n", routingKey, string(body))
	return nil
}
