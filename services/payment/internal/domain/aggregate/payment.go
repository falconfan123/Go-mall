package aggregate

import (
	"github.com/falconfan123/Go-mall/services/payment/internal/domain/entity"
)

// PaymentAggregate 支付单聚合根
type PaymentAggregate struct {
	payment *entity.Payment
}

// NewPaymentAggregate 创建支付单聚合根
func NewPaymentAggregate(
	paymentID string,
	preOrderID string,
	orderID string,
	userID int64,
	originalAmount int64,
	paidAmount int64,
	paymentMethod entity.PaymentMethod,
	payURL string,
	expireMinutes int,
) *PaymentAggregate {
	return &PaymentAggregate{
		payment: entity.NewPayment(
			paymentID,
			preOrderID,
			orderID,
			userID,
			originalAmount,
			paidAmount,
			paymentMethod,
			payURL,
			expireMinutes,
		),
	}
}

// LoadPayment 加载支付单到聚合根
func LoadPayment(payment *entity.Payment) *PaymentAggregate {
	return &PaymentAggregate{
		payment: payment,
	}
}

// GetPayment 获取支付单
func (p *PaymentAggregate) GetPayment() *entity.Payment {
	return p.payment
}

// Pay 支付成功
func (p *PaymentAggregate) Pay(transactionID string) error {
	return p.payment.Pay(transactionID)
}

// Cancel 取消支付
func (p *PaymentAggregate) Cancel() error {
	return p.payment.Cancel()
}

// Refund 退款
func (p *PaymentAggregate) Refund() error {
	return p.payment.Refund()
}

// IsExpired 检查是否过期
func (p *PaymentAggregate) IsExpired() bool {
	return p.payment.IsExpired()
}

// CanPay 检查是否可以支付
func (p *PaymentAggregate) CanPay() bool {
	return p.payment.CanPay()
}

// GetPaymentID 获取支付单ID
func (p *PaymentAggregate) GetPaymentID() string {
	return p.payment.PaymentID
}

// GetOrderID 获取订单ID
func (p *PaymentAggregate) GetOrderID() string {
	return p.payment.OrderID
}

// GetPreOrderID 获取预订单ID
func (p *PaymentAggregate) GetPreOrderID() string {
	return p.payment.PreOrderID
}

// GetUserID 获取用户ID
func (p *PaymentAggregate) GetUserID() int64 {
	return p.payment.UserID
}

// GetAmount 获取金额
func (p *PaymentAggregate) GetAmount() int64 {
	return p.payment.PaidAmount
}

// GetStatus 获取支付状态
func (p *PaymentAggregate) GetStatus() entity.PaymentStatus {
	return p.payment.Status
}

// GetPaymentMethod 获取支付渠道
func (p *PaymentAggregate) GetPaymentMethod() entity.PaymentMethod {
	return p.payment.PaymentMethod
}

// GetPayURL 获取支付链接
func (p *PaymentAggregate) GetPayURL() string {
	return p.payment.PayURL
}

// GetExpireTime 获取过期时间
func (p *PaymentAggregate) GetExpireTime() int64 {
	return p.payment.ExpireTime.Unix()
}
