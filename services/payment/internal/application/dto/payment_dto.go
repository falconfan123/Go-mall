package dto

// CreatePaymentReq 创建支付请求
type CreatePaymentReq struct {
	UserID        int64  `json:"userId"`
	OrderID       string `json:"orderId"`
	PaymentMethod int    `json:"paymentMethod"`
}

// CreatePaymentResp 创建支付响应
type CreatePaymentResp struct {
	Payment    *PaymentDTO `json:"payment"`
	StatusCode int64       `json:"statusCode"`
	StatusMsg  string      `json:"statusMsg"`
}

// GetPaymentReq 获取支付请求
type GetPaymentReq struct {
	PaymentID string `json:"paymentId"`
}

// GetPaymentResp 获取支付响应
type GetPaymentResp struct {
	Payment    *PaymentDTO `json:"payment"`
	StatusCode int64       `json:"statusCode"`
	StatusMsg  string      `json:"statusMsg"`
}

// ListPaymentsReq 支付列表请求
type ListPaymentsReq struct {
	UserID   int64 `json:"userId"`
	Status   *int  `json:"status"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// ListPaymentsResp 支付列表响应
type ListPaymentsResp struct {
	Payments   []*PaymentListItemDTO `json:"payments"`
	TotalCount int64                 `json:"totalCount"`
	StatusCode int64                 `json:"statusCode"`
	StatusMsg  string                `json:"statusMsg"`
}

// PaymentDTO 支付单DTO
type PaymentDTO struct {
	PaymentID      string `json:"paymentId"`
	PreOrderID     string `json:"preOrderId"`
	OrderID        string `json:"orderId"`
	UserID         int64  `json:"userId"`
	OriginalAmount int64  `json:"originalAmount"`
	PaidAmount     int64  `json:"paidAmount"`
	PaymentMethod  int    `json:"paymentMethod"`
	TransactionID  string `json:"transactionId"`
	PayURL         string `json:"payUrl"`
	Status         int    `json:"status"`
	ExpireTime     int64  `json:"expireTime"`
	PaidAt         *int64 `json:"paidAt,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
}

// PaymentListItemDTO 支付列表项DTO
type PaymentListItemDTO struct {
	PaymentID     string `json:"paymentId"`
	OrderID       string `json:"orderId"`
	PaidAmount    int64  `json:"paidAmount"`
	Status        int    `json:"status"`
	PaymentMethod int    `json:"paymentMethod"`
	CreatedAt     int64  `json:"createdAt"`
}
