package valueobject

import "errors"

// Quantity 商品数量值对象
type Quantity int32

var (
	ErrInvalidQuantity     = errors.New("quantity must be greater than 0")
	ErrMaxQuantityExceeded = errors.New("maximum quantity per item is 99")
)

const MaxQuantity = 99

// NewQuantity 创建数量值对象
func NewQuantity(value int32) (Quantity, error) {
	if value <= 0 {
		return 0, ErrInvalidQuantity
	}
	if value > MaxQuantity {
		return 0, ErrMaxQuantityExceeded
	}
	return Quantity(value), nil
}

// Value 获取数量值
func (q Quantity) Value() int32 {
	return int32(q)
}

// Add 增加数量
func (q Quantity) Add(delta int32) (Quantity, error) {
	newVal := q + Quantity(delta)
	if newVal > MaxQuantity {
		return 0, ErrMaxQuantityExceeded
	}
	return newVal, nil
}

// Subtract 减少数量
func (q Quantity) Subtract(delta int32) (Quantity, error) {
	newVal := q - Quantity(delta)
	if newVal < 0 {
		return 0, ErrInvalidQuantity
	}
	return newVal, nil
}
