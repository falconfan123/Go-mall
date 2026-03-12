package entity

import (
	"errors"
	"time"
)

// OrderStatus 订单状态
type OrderStatus int

const (
	OrderStatusPending   OrderStatus = 1 // 待支付
	OrderStatusPaid      OrderStatus = 2 // 已支付
	OrderStatusShipped   OrderStatus = 3 // 已发货
	OrderStatusCompleted OrderStatus = 4 // 已完成
	OrderStatusCanceled  OrderStatus = 5 // 已取消
)

// PaymentStatus 支付状态
type PaymentStatus int

const (
	PaymentStatusUnpaid   PaymentStatus = 0 // 未支付
	PaymentStatusPaying   PaymentStatus = 1 // 支付中
	PaymentStatusPaid     PaymentStatus = 2 // 已支付
	PaymentStatusRefunded PaymentStatus = 3 // 已退款
)

var (
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
	ErrOrderNotPaid         = errors.New("order not paid")
	ErrOrderAlreadyPaid     = errors.New("order already paid")
)

// Order 订单实体
type Order struct {
	OrderID    string // 订单ID
	PreOrderID string // 预订单ID
	UserID     int64  // 用户ID
	CouponID   string // 优惠券ID

	// 支付信息
	PaymentMethod int        // 支付方式
	TransactionID string     // 交易流水号
	PaidAt        *time.Time // 支付时间

	// 金额信息
	OriginalAmount int64 // 原始金额（分）
	DiscountAmount int64 // 优惠金额（分）
	PayableAmount  int64 // 应付金额（分）
	PaidAmount     int64 // 实收金额（分）

	// 状态
	OrderStatus   OrderStatus   // 订单状态
	PaymentStatus PaymentStatus // 支付状态
	Reason        string        // 取消原因

	// 时间
	ExpireTime time.Time // 过期时间
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间

	// 关联
	Items   []*OrderItem  // 订单项
	Address *OrderAddress // 地址快照
}

// NewOrder 创建订单
func NewOrder(
	orderID string,
	preOrderID string,
	userID int64,
	couponID string,
	originalAmount int64,
	discountAmount int64,
	payableAmount int64,
	expireMinutes int,
) *Order {
	return &Order{
		OrderID:        orderID,
		PreOrderID:     preOrderID,
		UserID:         userID,
		CouponID:       couponID,
		OriginalAmount: originalAmount,
		DiscountAmount: discountAmount,
		PayableAmount:  payableAmount,
		PaidAmount:     0,
		OrderStatus:    OrderStatusPending,
		PaymentStatus:  PaymentStatusUnpaid,
		ExpireTime:     time.Now().Add(time.Duration(expireMinutes) * time.Minute),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Items:          make([]*OrderItem, 0),
	}
}

// AddItem 添加订单项
func (o *Order) AddItem(item *OrderItem) {
	o.Items = append(o.Items, item)
}

// SetAddress 设置地址快照
func (o *Order) SetAddress(address *OrderAddress) {
	o.Address = address
}

// Pay 支付
func (o *Order) Pay(paymentMethod int, transactionID string) error {
	if o.OrderStatus != OrderStatusPending {
		return ErrInvalidOrderStatus
	}
	if o.PaymentStatus == PaymentStatusPaid {
		return ErrOrderAlreadyPaid
	}

	now := time.Now()
	o.PaymentMethod = paymentMethod
	o.TransactionID = transactionID
	o.PaidAt = &now
	o.PaidAmount = o.PayableAmount
	o.PaymentStatus = PaymentStatusPaid
	o.OrderStatus = OrderStatusPaid
	o.UpdatedAt = time.Now()
	return nil
}

// Ship 发货
func (o *Order) Ship() error {
	if o.OrderStatus != OrderStatusPaid {
		return ErrOrderNotPaid
	}
	o.OrderStatus = OrderStatusShipped
	o.UpdatedAt = time.Now()
	return nil
}

// Complete 完成
func (o *Order) Complete() error {
	if o.OrderStatus != OrderStatusShipped {
		return ErrInvalidOrderStatus
	}
	o.OrderStatus = OrderStatusCompleted
	o.UpdatedAt = time.Now()
	return nil
}

// Cancel 取消
func (o *Order) Cancel(reason string) error {
	if o.OrderStatus == OrderStatusPaid || o.OrderStatus == OrderStatusShipped || o.OrderStatus == OrderStatusCompleted {
		return ErrInvalidOrderStatus
	}
	o.OrderStatus = OrderStatusCanceled
	o.Reason = reason
	o.UpdatedAt = time.Now()
	return nil
}

// Refund 退款
func (o *Order) Refund() error {
	if o.PaymentStatus != PaymentStatusPaid {
		return ErrOrderNotPaid
	}
	o.PaymentStatus = PaymentStatusRefunded
	o.OrderStatus = OrderStatusCanceled
	o.UpdatedAt = time.Now()
	return nil
}

// IsExpired 是否过期
func (o *Order) IsExpired() bool {
	return time.Now().After(o.ExpireTime)
}

// CanCancel 是否可取消
func (o *Order) CanCancel() bool {
	return o.OrderStatus == OrderStatusPending
}

// OrderItem 订单项实体
type OrderItem struct {
	OrderID     string // 订单ID
	ProductID   int64  // 商品ID
	Quantity    int    // 数量
	Price       int64  // 单价（分）
	ProductName string // 商品名称
	ProductDesc string // 商品描述
	CreatedAt   time.Time
}

// NewOrderItem 创建订单项
func NewOrderItem(
	orderID string,
	productID int64,
	quantity int,
	price int64,
	productName string,
	productDesc string,
) *OrderItem {
	return &OrderItem{
		OrderID:     orderID,
		ProductID:   productID,
		Quantity:    quantity,
		Price:       price,
		ProductName: productName,
		ProductDesc: productDesc,
		CreatedAt:   time.Now(),
	}
}

// TotalPrice 计算总价
func (i *OrderItem) TotalPrice() int64 {
	return i.Price * int64(i.Quantity)
}

// OrderAddress 订单地址快照实体
type OrderAddress struct {
	OrderID         string // 订单ID
	AddressID       int64  // 地址ID
	RecipientName   string // 收件人姓名
	PhoneNumber     string // 联系电话
	Province        string // 省份
	City            string // 城市
	DetailedAddress string // 详细地址
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewOrderAddress 创建订单地址快照
func NewOrderAddress(
	orderID string,
	addressID int64,
	recipientName string,
	phoneNumber string,
	province string,
	city string,
	detailedAddress string,
) *OrderAddress {
	now := time.Now()
	return &OrderAddress{
		OrderID:         orderID,
		AddressID:       addressID,
		RecipientName:   recipientName,
		PhoneNumber:     phoneNumber,
		Province:        province,
		City:            city,
		DetailedAddress: detailedAddress,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
