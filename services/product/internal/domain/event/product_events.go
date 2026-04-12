package event

import "time"

// BaseEvent 领域事件基类
type BaseEvent struct {
	EventID    string    // 事件ID
	EventType  string    // 事件类型
	OccurredAt time.Time // 事件发生时间
}

// ProductCreatedEvent 商品创建事件
type ProductCreatedEvent struct {
	BaseEvent
	ProductID   int64  // 商品ID
	ProductName string // 商品名称
	Price       int64  // 商品价格（分）
	Stock       int64  // 商品库存
}

// ProductUpdatedEvent 商品更新事件
type ProductUpdatedEvent struct {
	BaseEvent
	ProductID   int64  // 商品ID
	ProductName string // 商品名称
	Price       int64  // 更新后的价格（分）
}

// ProductStockChangedEvent 商品库存变化事件
type ProductStockChangedEvent struct {
	BaseEvent
	ProductID     int64 // 商品ID
	OldStock      int64 // 原库存
	NewStock      int64 // 新库存
	ChangedAmount int64 // 变化量（正数增加，负数减少）
}

// ProductDeletedEvent 商品删除事件
type ProductDeletedEvent struct {
	BaseEvent
	ProductID int64 // 商品ID
}
