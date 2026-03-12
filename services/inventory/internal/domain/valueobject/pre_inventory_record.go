package valueobject

import (
	"errors"
	"time"
)

// PreInventoryRecord 预扣库存记录值对象
type PreInventoryRecord struct {
	ProductID  int64     // 商品ID
	Quantity   int64     // 预扣数量
	PreOrderID string    // 预订单ID
	UserID     int64     // 用户ID
	ExpireTime time.Time // 过期时间
}

var (
	ErrInvalidProductID  = errors.New("product ID cannot be negative")
	ErrInvalidQuantity   = errors.New("quantity must be greater than 0")
	ErrEmptyPreOrderID   = errors.New("pre order ID cannot be empty")
	ErrInvalidExpireTime = errors.New("expire time must be in the future")
)

// NewPreInventoryRecord 创建预扣库存记录
func NewPreInventoryRecord(
	productID int64,
	quantity int64,
	preOrderID string,
	userID int64,
	expireTime time.Time,
) (PreInventoryRecord, error) {
	if productID < 0 {
		return PreInventoryRecord{}, ErrInvalidProductID
	}
	if quantity <= 0 {
		return PreInventoryRecord{}, ErrInvalidQuantity
	}
	if preOrderID == "" {
		return PreInventoryRecord{}, ErrEmptyPreOrderID
	}
	if expireTime.Before(time.Now()) {
		return PreInventoryRecord{}, ErrInvalidExpireTime
	}

	return PreInventoryRecord{
		ProductID:  productID,
		Quantity:   quantity,
		PreOrderID: preOrderID,
		UserID:     userID,
		ExpireTime: expireTime,
	}, nil
}

// IsExpired 判断预扣记录是否过期
func (r PreInventoryRecord) IsExpired() bool {
	return r.ExpireTime.Before(time.Now())
}

// Equals 判断两个预扣记录是否相等
func (r PreInventoryRecord) Equals(other PreInventoryRecord) bool {
	return r.PreOrderID == other.PreOrderID && r.ProductID == other.ProductID
}
