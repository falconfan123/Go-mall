package entity

import (
	"errors"
	"time"
)

// PaymentStatus 支付状态
type PaymentStatus int

const (
	PaymentStatusUnpaid   PaymentStatus = 0 // 未支付
	PaymentStatusPending  PaymentStatus = 1 // 待支付
	PaymentStatusPaid     PaymentStatus = 2 // 已支付
	PaymentStatusRefunded PaymentStatus = 3 // 已退款
	PaymentStatusFailed   PaymentStatus = 4 // 支付失败
)

var (
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentExpired       = errors.New("payment expired")
)

// Payment 支付单实体
type Payment struct {
	PaymentID      string // 支付单ID
	PreOrderID     string // 预订单ID
	OrderID        string // 订单ID
	UserID         int64  // 用户ID
	OriginalAmount int64  // 订单原价（分）
	PaidAmount     int64  // 实付金额（分）

	// 支付信息
	PaymentMethod PaymentMethod // 支付渠道
	TransactionID string        // 支付平台交易号
	PayURL        string        // 支付跳转链接
	ExpireTime    time.Time     // 支付链接过期时间

	// 状态
	Status PaymentStatus // 支付状态

	// 时间
	PaidAt    *time.Time // 支付成功时间
	CreatedAt time.Time  // 创建时间
	UpdatedAt time.Time  // 更新时间
}

// PaymentMethod 支付渠道
type PaymentMethod string

const (
	PaymentMethodAlipay  PaymentMethod = "alipay"  // 支付宝
	PaymentMethodWechat  PaymentMethod = "wx_pay"  // 微信支付
	PaymentMethodUnknown PaymentMethod = "unknown" // 未知
)

// NewPayment 创建支付单
func NewPayment(
	paymentID string,
	preOrderID string,
	orderID string,
	userID int64,
	originalAmount int64,
	paidAmount int64,
	paymentMethod PaymentMethod,
	payURL string,
	expireMinutes int,
) *Payment {
	return &Payment{
		PaymentID:      paymentID,
		PreOrderID:     preOrderID,
		OrderID:        orderID,
		UserID:         userID,
		OriginalAmount: originalAmount,
		PaidAmount:     paidAmount,
		PaymentMethod:  paymentMethod,
		PayURL:         payURL,
		ExpireTime:     time.Now().Add(time.Duration(expireMinutes) * time.Minute),
		Status:         PaymentStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Pay 支付成功
func (p *Payment) Pay(transactionID string) error {
	if p.Status != PaymentStatusPending {
		return ErrInvalidPaymentStatus
	}

	now := time.Now()
	p.TransactionID = transactionID
	p.PaidAt = &now
	p.Status = PaymentStatusPaid
	p.UpdatedAt = time.Now()
	return nil
}

// Cancel 取消支付
func (p *Payment) Cancel() error {
	if p.Status == PaymentStatusPaid {
		return ErrInvalidPaymentStatus
	}
	p.Status = PaymentStatusFailed
	p.UpdatedAt = time.Now()
	return nil
}

// Refund 退款
func (p *Payment) Refund() error {
	if p.Status != PaymentStatusPaid {
		return ErrInvalidPaymentStatus
	}
	p.Status = PaymentStatusRefunded
	p.UpdatedAt = time.Now()
	return nil
}

// IsExpired 是否过期
func (p *Payment) IsExpired() bool {
	return time.Now().After(p.ExpireTime)
}

// CanPay 是否可以支付
func (p *Payment) CanPay() bool {
	return p.Status == PaymentStatusPending && !p.IsExpired()
}

// String 转换为字符串
func (m PaymentMethod) String() string {
	switch m {
	case PaymentMethodAlipay:
		return "alipay"
	case PaymentMethodWechat:
		return "wx_pay"
	default:
		return "unknown"
	}
}

// FromString 从字符串创建
func PaymentMethodFromString(s string) PaymentMethod {
	switch s {
	case "alipay":
		return PaymentMethodAlipay
	case "wx_pay":
		return PaymentMethodWechat
	default:
		return PaymentMethodUnknown
	}
}
