package event

import (
	"github.com/falconfan123/Go-mall/services/product/internal/domain/event"
)

// ProductEventPublisher 商品事件发布器接口，定义领域事件的发布能力
type ProductEventPublisher interface {
	// PublishProductCreated 发布商品创建事件
	PublishProductCreated(evt *event.ProductCreatedEvent) error

	// PublishProductUpdated 发布商品更新事件
	PublishProductUpdated(evt *event.ProductUpdatedEvent) error

	// PublishProductStockChanged 发布商品库存变化事件
	PublishProductStockChanged(evt *event.ProductStockChangedEvent) error

	// PublishProductDeleted 发布商品删除事件
	PublishProductDeleted(evt *event.ProductDeletedEvent) error
}
