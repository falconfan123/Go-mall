package valueobject

import "errors"

// Price 价格值对象，单位：分
type Price int64

var (
	ErrInvalidPrice = errors.New("price cannot be negative")
)

// NewPrice 创建价格值对象
func NewPrice(value int64) (Price, error) {
	if value < 0 {
		return 0, ErrInvalidPrice
	}
	return Price(value), nil
}

// Value 获取价格值
func (p Price) Value() int64 {
	return int64(p)
}

// Add 价格相加
func (p Price) Add(other Price) Price {
	return p + other
}

// Subtract 价格相减
func (p Price) Subtract(other Price) Price {
	if p < other {
		return 0
	}
	return p - other
}

// Multiply 价格乘以数量
func (p Price) Multiply(quantity int64) Price {
	return p * Price(quantity)
}

// ToYuan 转换为元（保留两位小数）
func (p Price) ToYuan() float64 {
	return float64(p) / 100.0
}
