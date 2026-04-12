package valueobject

import "time"

// Money 金额值对象
type Money struct {
	Amount   int64  // 金额（分）
	Currency string // 货币类型
}

// NewMoney 创建金额
func NewMoney(amount int64) *Money {
	return &Money{
		Amount:   amount,
		Currency: "CNY",
	}
}

// Add 相加
func (m *Money) Add(other *Money) *Money {
	return &Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}
}

// Subtract 相减
func (m *Money) Subtract(other *Money) *Money {
	amount := m.Amount - other.Amount
	if amount < 0 {
		amount = 0
	}
	return &Money{
		Amount:   amount,
		Currency: m.Currency,
	}
}

// ToYuan 转换为元
func (m *Money) ToYuan() float64 {
	return float64(m.Amount) / 100.0
}

// PaymentMethod 支付方式
type PaymentMethod int

const (
	PaymentMethodWechat PaymentMethod = 1 // 微信
	PaymentMethodAlipay PaymentMethod = 2 // 支付宝
)

// OrderID 订单ID值对象
type OrderID struct {
	Value string
}

// NewOrderID 创建订单ID
func NewOrderID() *OrderID {
	return &OrderID{
		Value: generateOrderID(),
	}
}

func generateOrderID() string {
	return time.Now().Format("20060102150405")
}
