package aggregate

import (
	"errors"

	"github.com/falconfan123/Go-mall/services/order/internal/domain/entity"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInconsistentAmount = errors.New("inconsistent amount")
)

// OrderAggregate 订单聚合根
type OrderAggregate struct {
	order *entity.Order
}

// NewOrderAggregate 创建订单聚合根
func NewOrderAggregate(
	orderID string,
	preOrderID string,
	userID int64,
	couponID string,
	originalAmount int64,
	discountAmount int64,
	payableAmount int64,
	expireMinutes int,
) *OrderAggregate {
	return &OrderAggregate{
		order: entity.NewOrder(orderID, preOrderID, userID, couponID, originalAmount, discountAmount, payableAmount, expireMinutes),
	}
}

// LoadOrder 加载订单到聚合根
func LoadOrder(order *entity.Order) *OrderAggregate {
	return &OrderAggregate{
		order: order,
	}
}

// GetOrder 获取订单
func (o *OrderAggregate) GetOrder() *entity.Order {
	return o.order
}

// AddItem 添加订单项
func (o *OrderAggregate) AddItem(
	productID int64,
	quantity int,
	price int64,
	productName string,
	productDesc string,
) error {
	item := entity.NewOrderItem(
		o.order.OrderID,
		productID,
		quantity,
		price,
		productName,
		productDesc,
	)
	o.order.AddItem(item)
	return nil
}

// SetAddress 设置地址快照
func (o *OrderAggregate) SetAddress(
	addressID int64,
	recipientName string,
	phoneNumber string,
	province string,
	city string,
	detailedAddress string,
) {
	address := entity.NewOrderAddress(
		o.order.OrderID,
		addressID,
		recipientName,
		phoneNumber,
		province,
		city,
		detailedAddress,
	)
	o.order.SetAddress(address)
}

// Pay 支付
func (o *OrderAggregate) Pay(paymentMethod int, transactionID string) error {
	return o.order.Pay(paymentMethod, transactionID)
}

// Ship 发货
func (o *OrderAggregate) Ship() error {
	return o.order.Ship()
}

// Complete 完成
func (o *OrderAggregate) Complete() error {
	return o.order.Complete()
}

// Cancel 取消
func (o *OrderAggregate) Cancel(reason string) error {
	return o.order.Cancel(reason)
}

// Refund 退款
func (o *OrderAggregate) Refund() error {
	return o.order.Refund()
}

// IsExpired 检查是否过期
func (o *OrderAggregate) IsExpired() bool {
	return o.order.IsExpired()
}

// GetOrderID 获取订单ID
func (o *OrderAggregate) GetOrderID() string {
	return o.order.OrderID
}

// GetUserID 获取用户ID
func (o *OrderAggregate) GetUserID() int64 {
	return o.order.UserID
}

// GetPreOrderID 获取预订单ID
func (o *OrderAggregate) GetPreOrderID() string {
	return o.order.PreOrderID
}

// GetTotalAmount 获取总金额
func (o *OrderAggregate) GetTotalAmount() int64 {
	return o.order.PayableAmount
}

// GetStatus 获取订单状态
func (o *OrderAggregate) GetStatus() entity.OrderStatus {
	return o.order.OrderStatus
}

// GetPaymentStatus 获取支付状态
func (o *OrderAggregate) GetPaymentStatus() entity.PaymentStatus {
	return o.order.PaymentStatus
}
