package valueobject

import "time"

// Address 收货地址值对象
type Address struct {
	ID        int64  // 地址ID
	UserID    int64  // 用户ID
	Name      string // 收货人姓名
	Phone     string // 联系电话
	Province  string // 省份
	City      string // 城市
	District  string // 区县
	Detail    string // 详细地址
	ZipCode   string // 邮编
	IsDefault bool   // 是否默认
}

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

// Multiply 乘法
func (m *Money) Multiply(factor int64) *Money {
	return &Money{
		Amount:   m.Amount * factor,
		Currency: m.Currency,
	}
}

// ToYuan 转换为元
func (m *Money) ToYuan() float64 {
	return float64(m.Amount) / 100.0
}

// OrderID 订单ID值对象
type OrderID struct {
	Value string // 订单ID
}

// PreOrderID 预订单ID值对象
type PreOrderID struct {
	Value string // 预订单ID
}

// NewPreOrderID 创建预订单ID
func NewPreOrderID(userID int64) *PreOrderID {
	return &PreOrderID{
		Value: generatePreOrderID(userID),
	}
}

func generatePreOrderID(userID int64) string {
	return time.Now().Format("20060102150405") + "_" + string(rune(userID))
}
