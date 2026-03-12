package valueobject

import "errors"

// Stock 库存值对象
type Stock int64

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidStock      = errors.New("stock cannot be negative")
)

// NewStock 创建库存值对象
func NewStock(value int64) (Stock, error) {
	if value < 0 {
		return 0, ErrInvalidStock
	}
	return Stock(value), nil
}

// Value 获取库存值
func (s Stock) Value() int64 {
	return int64(s)
}

// Adjust 调整库存，正数增加，负数减少
func (s Stock) Adjust(quantity int64) (Stock, error) {
	newStock := s + Stock(quantity)
	if newStock < 0 {
		return 0, ErrInsufficientStock
	}
	return newStock, nil
}

// IsAvailable 库存是否可用
func (s Stock) IsAvailable(quantity int64) bool {
	return s >= Stock(quantity)
}
