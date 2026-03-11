package valueobject

import "errors"

// Stock 库存数量值对象，不可变
type Stock int64

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidStock      = errors.New("stock cannot be negative")
)

// NewStock 创建库存数量值对象
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

// Add 增加库存
func (s Stock) Add(quantity int64) (Stock, error) {
	newStock := s + Stock(quantity)
	if newStock < 0 {
		return 0, ErrInvalidStock
	}
	return newStock, nil
}

// Subtract 扣减库存
func (s Stock) Subtract(quantity int64) (Stock, error) {
	if s < Stock(quantity) {
		return 0, ErrInsufficientStock
	}
	return s - Stock(quantity), nil
}

// IsAvailable 判断库存是否足够
func (s Stock) IsAvailable(quantity int64) bool {
	return s >= Stock(quantity)
}
