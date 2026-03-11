package event

import (
	"github.com/falconfan123/Go-mall/services/users/internal/domain/event"
)

// EventPublisher 领域事件发布器接口
type EventPublisher interface {
	// PublishUserRegistered 发布用户注册事件
	PublishUserRegistered(e *event.UserRegisteredEvent) error
	// PublishUserLoggedIn 发布用户登录事件
	PublishUserLoggedIn(e *event.UserLoggedInEvent) error
	// PublishUserLoggedOut 发布用户登出事件
	PublishUserLoggedOut(e *event.UserLoggedOutEvent) error
	// PublishUserInfoUpdated 发布用户信息更新事件
	PublishUserInfoUpdated(e *event.UserInfoUpdatedEvent) error
	// PublishAddressAdded 发布地址添加事件
	PublishAddressAdded(e *event.AddressAddedEvent) error
	// PublishAddressUpdated 发布地址更新事件
	PublishAddressUpdated(e *event.AddressUpdatedEvent) error
	// PublishAddressDeleted 发布地址删除事件
	PublishAddressDeleted(e *event.AddressDeletedEvent) error
}
