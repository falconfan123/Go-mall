package event

import (
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/event"
)

// InventoryEventPublisher 库存事件发布器接口
type InventoryEventPublisher interface {
	// PublishInventoryPreDecreased 发布库存预扣减事件
	PublishInventoryPreDecreased(evt *event.InventoryPreDecreasedEvent) error

	// PublishInventoryDecreased 发布库存扣减事件
	PublishInventoryDecreased(evt *event.InventoryDecreasedEvent) error

	// PublishInventoryPreReturned 发布预扣库存退还事件
	PublishInventoryPreReturned(evt *event.InventoryPreReturnedEvent) error

	// PublishInventoryReturned 发布库存退还事件
	PublishInventoryReturned(evt *event.InventoryReturnedEvent) error

	// PublishInventoryUpdated 发布库存更新事件
	PublishInventoryUpdated(evt *event.InventoryUpdatedEvent) error
}
