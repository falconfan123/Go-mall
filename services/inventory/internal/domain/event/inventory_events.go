package event

import "time"

// BaseEvent 领域事件基类
type BaseEvent struct {
	EventID    string    // 事件ID
	EventType  string    // 事件类型
	OccurredAt time.Time // 事件发生时间
}

// InventoryPreDecreasedEvent 库存预扣减事件
type InventoryPreDecreasedEvent struct {
	BaseEvent
	ProductID  int64  // 商品ID
	Quantity   int64  // 数量
	PreOrderID string // 预订单ID
	UserID     int64  // 用户ID
}

// InventoryDecreasedEvent 库存扣减事件
type InventoryDecreasedEvent struct {
	BaseEvent
	ProductID int64  // 商品ID
	Quantity  int64  // 数量
	OrderID   string // 订单ID
}

// InventoryPreReturnedEvent 预扣库存退还事件
type InventoryPreReturnedEvent struct {
	BaseEvent
	ProductID  int64  // 商品ID
	Quantity   int64  // 数量
	PreOrderID string // 预订单ID
}

// InventoryReturnedEvent 库存退还事件
type InventoryReturnedEvent struct {
	BaseEvent
	ProductID int64  // 商品ID
	Quantity  int64  // 数量
	OrderID   string // 订单ID
}

// InventoryUpdatedEvent 库存更新事件
type InventoryUpdatedEvent struct {
	BaseEvent
	ProductID     int64 // 商品ID
	OldStock      int64 // 原库存
	NewStock      int64 // 新库存
	ChangedAmount int64 // 变化量
}
