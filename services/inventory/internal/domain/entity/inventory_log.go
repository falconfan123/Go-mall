package entity

import (
	"time"
)

// InventoryLog 库存变更记录实体
type InventoryLog struct {
	ID        int64     // 记录ID
	ProductID int64     // 商品ID
	Quantity  int64     // 变更数量（正数为增加，负数为减少）
	Type      LogType   // 变更类型
	OrderID   string    // 关联订单ID
	Remark    string    // 备注
	CreatedAt time.Time // 创建时间
}

// LogType 库存变更类型
type LogType int

const (
	LogTypeInitial         LogType = 1 // 初始化
	LogTypePreDecrease     LogType = 2 // 预扣减
	LogTypeConfirmDecrease LogType = 3 // 确认扣减
	LogTypeReturn          LogType = 4 // 退还
	LogTypeDirectDecrease  LogType = 5 // 直接扣减
	LogTypeAdjust          LogType = 6 // 调整
)

// NewInventoryLog 创建库存变更记录
func NewInventoryLog(
	productID int64,
	quantity int64,
	logType LogType,
	orderID string,
	remark string,
) *InventoryLog {
	return &InventoryLog{
		ProductID: productID,
		Quantity:  quantity,
		Type:      logType,
		OrderID:   orderID,
		Remark:    remark,
		CreatedAt: time.Now(),
	}
}

// IsIncrease 是否为增加操作
func (l *InventoryLog) IsIncrease() bool {
	return l.Quantity > 0
}

// IsDecrease 是否为减少操作
func (l *InventoryLog) IsDecrease() bool {
	return l.Quantity < 0
}
