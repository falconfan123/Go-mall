package order

// OrderStatus 订单状态枚举
type OrderStatus int32

const (
	OrderStatusUnspecified    OrderStatus = 0
	OrderStatusCreated        OrderStatus = 1 // 创建
	OrderStatusPendingPayment OrderStatus = 2 // 待支付
	OrderStatusPaid           OrderStatus = 3 // 已支付
	OrderStatusCompleted      OrderStatus = 4 // 已完成
	OrderStatusCancelled      OrderStatus = 5 // 已取消
	OrderStatusClosed         OrderStatus = 6 // 已关闭（超时）
	OrderStatusRefund         OrderStatus = 7 // 退款
)

// PaymentStatus 支付状态枚举
type PaymentStatus int32

const (
	PaymentStatusUnspecified PaymentStatus = 0
	PaymentStatusNotPaid     PaymentStatus = 1 // 未支付
	PaymentStatusPaying      PaymentStatus = 2 // 支付中
	PaymentStatusPaid        PaymentStatus = 3 // 已支付
	PaymentStatusExpired     PaymentStatus = 4 // 已过期
	PaymentStatusRefund      PaymentStatus = 5 // 退款
)

// PaymentMethod 支付方式枚举
type PaymentMethod int32

const (
	PaymentMethodUnspecified PaymentMethod = 0
	PaymentMethodWechatPay   PaymentMethod = 1
	PaymentMethodAlipay      PaymentMethod = 2
)

// PaymentResult 支付结果
type PaymentResult struct {
	TransactionId string // 支付平台流水号
	PaidAmount    int64  // 实际支付金额（分）
	PaidAt        int64  // 支付时间戳
}
