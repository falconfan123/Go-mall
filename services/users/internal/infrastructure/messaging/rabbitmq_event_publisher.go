package messaging

import (
	"encoding/json"
	"github.com/falconfan123/Go-mall/services/users/internal/application/event"
	domainevent "github.com/falconfan123/Go-mall/services/users/internal/domain/event"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-queue/rabbitmq"
)

// RabbitMQEventPublisher RabbitMQ事件发布器实现
type RabbitMQEventPublisher struct {
	producer *rabbitmq.Producer
}

// NewRabbitMQEventPublisher 创建RabbitMQ事件发布器
func NewRabbitMQEventPublisher(producer *rabbitmq.Producer) event.EventPublisher {
	return &RabbitMQEventPublisher{
		producer: producer,
	}
}

// PublishUserRegistered 发布用户注册事件
func (p *RabbitMQEventPublisher) PublishUserRegistered(e *domainevent.UserRegisteredEvent) error {
	return p.publish("user.event.registered", e)
}

// PublishUserLoggedIn 发布用户登录事件
func (p *RabbitMQEventPublisher) PublishUserLoggedIn(e *domainevent.UserLoggedInEvent) error {
	return p.publish("user.event.logged_in", e)
}

// PublishUserLoggedOut 发布用户登出事件
func (p *RabbitMQEventPublisher) PublishUserLoggedOut(e *domainevent.UserLoggedOutEvent) error {
	return p.publish("user.event.logged_out", e)
}

// PublishUserInfoUpdated 发布用户信息更新事件
func (p *RabbitMQEventPublisher) PublishUserInfoUpdated(e *domainevent.UserInfoUpdatedEvent) error {
	return p.publish("user.event.info_updated", e)
}

// PublishAddressAdded 发布地址添加事件
func (p *RabbitMQEventPublisher) PublishAddressAdded(e *domainevent.AddressAddedEvent) error {
	return p.publish("user.event.address_added", e)
}

// PublishAddressUpdated 发布地址更新事件
func (p *RabbitMQEventPublisher) PublishAddressUpdated(e *domainevent.AddressUpdatedEvent) error {
	return p.publish("user.event.address_updated", e)
}

// PublishAddressDeleted 发布地址删除事件
func (p *RabbitMQEventPublisher) PublishAddressDeleted(e *domainevent.AddressDeletedEvent) error {
	return p.publish("user.event.address_deleted", e)
}

// publish 通用发布方法
func (p *RabbitMQEventPublisher) publish(routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		logx.Errorw("marshal event failed", logx.Field("err", err), logx.Field("event", event))
		return err
	}

	err = p.producer.Publish(routingKey, body)
	if err != nil {
		logx.Errorw("publish event failed", logx.Field("err", err), logx.Field("routing_key", routingKey))
		return err
	}

	logx.Infow("event published", logx.Field("routing_key", routingKey), logx.Field("event_id", getEventID(event)))
	return nil
}

// getEventID 从事件中获取事件ID
func getEventID(event interface{}) string {
	switch e := event.(type) {
	case *domainevent.UserRegisteredEvent:
		return e.EventID
	case *domainevent.UserLoggedInEvent:
		return e.EventID
	case *domainevent.UserLoggedOutEvent:
		return e.EventID
	case *domainevent.UserInfoUpdatedEvent:
		return e.EventID
	case *domainevent.AddressAddedEvent:
		return e.EventID
	case *domainevent.AddressUpdatedEvent:
		return e.EventID
	case *domainevent.AddressDeletedEvent:
		return e.EventID
	default:
		return ""
	}
}
