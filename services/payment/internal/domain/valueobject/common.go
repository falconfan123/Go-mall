package valueobject

import "time"

// Money 金额
type Money struct {
	amount int64 // 单位：分
}

// NewMoney 创建金额
func NewMoney(amount int64) *Money {
	return &Money{amount: amount}
}

// Amount 获取金额
func (m *Money) Amount() int64 {
	return m.amount
}

// Add 相加
func (m *Money) Add(other *Money) *Money {
	return &Money{amount: m.amount + other.amount}
}

// Subtract 相减
func (m *Money) Subtract(other *Money) *Money {
	return &Money{amount: m.amount - other.amount}
}

// PaymentID 支付单ID
type PaymentID struct {
	value string
}

// NewPaymentID 创建支付单ID
func NewPaymentID(value string) *PaymentID {
	return &PaymentID{value: value}
}

// Value 获取值
func (p *PaymentID) Value() string {
	return p.value
}

// PreOrderID 预订单ID
type PreOrderID struct {
	value string
}

// NewPreOrderID 创建预订单ID
func NewPreOrderID(value string) *PreOrderID {
	return &PreOrderID{value: value}
}

// Value 获取值
func (p *PreOrderID) Value() string {
	return p.value
}

// OrderID 订单ID
type OrderID struct {
	value string
}

// NewOrderID 创建订单ID
func NewOrderID(value string) *OrderID {
	return &OrderID{value: value}
}

// Value 获取值
func (o *OrderID) Value() string {
	return o.value
}

// PaymentTime 支付时间
type PaymentTime struct {
	time time.Time
}

// NewPaymentTime 创建支付时间
func NewPaymentTime(t time.Time) *PaymentTime {
	return &PaymentTime{time: t}
}

// Value 获取值
func (p *PaymentTime) Value() time.Time {
	return p.time
}

// ExpireTime 过期时间
type ExpireTime struct {
	time time.Time
}

// NewExpireTime 创建过期时间
func NewExpireTime(t time.Time) *ExpireTime {
	return &ExpireTime{time: t}
}

// Value 获取值
func (e *ExpireTime) Value() time.Time {
	return e.time
}

// IsExpired 是否过期
func (e *ExpireTime) IsExpired() bool {
	return time.Now().After(e.time)
}
