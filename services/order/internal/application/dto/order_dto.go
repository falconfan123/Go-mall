package dto

// CreateOrderReq 创建订单请求
type CreateOrderReq struct {
	PreOrderID     string           `json:"preOrderId"`
	UserID         int64            `json:"userId"`
	CouponID       string           `json:"couponId"`
	OriginalAmount int64            `json:"originalAmount"`
	DiscountAmount int64            `json:"discountAmount"`
	PayableAmount  int64            `json:"payableAmount"`
	Items          []*OrderItemReq  `json:"items"`
	Address        *OrderAddressReq `json:"address"`
}

// OrderItemReq 订单项请求
type OrderItemReq struct {
	ProductID   int64  `json:"productId"`
	ProductName string `json:"productName"`
	ProductDesc string `json:"productDesc"`
	Quantity    int    `json:"quantity"`
	Price       int64  `json:"price"`
}

// OrderAddressReq 订单地址请求
type OrderAddressReq struct {
	AddressID       int64  `json:"addressId"`
	RecipientName   string `json:"recipientName"`
	PhoneNumber     string `json:"phoneNumber"`
	Province        string `json:"province"`
	City            string `json:"city"`
	DetailedAddress string `json:"detailedAddress"`
}

// CreateOrderResp 创建订单响应
type CreateOrderResp struct {
	OrderID    string `json:"orderId"`
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// GetOrderReq 获取订单请求
type GetOrderReq struct {
	OrderID string `json:"orderId"`
}

// GetOrderResp 获取订单响应
type GetOrderResp struct {
	Order      *OrderDTO `json:"order"`
	StatusCode int64     `json:"statusCode"`
	StatusMsg  string    `json:"statusMsg"`
}

// ListOrdersReq 订单列表请求
type ListOrdersReq struct {
	UserID   int64 `json:"userId"`
	Status   *int  `json:"status"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// ListOrdersResp 订单列表响应
type ListOrdersResp struct {
	Orders     []*OrderListItemDTO `json:"orders"`
	TotalCount int64               `json:"totalCount"`
	StatusCode int64               `json:"statusCode"`
	StatusMsg  string              `json:"statusMsg"`
}

// CancelOrderReq 取消订单请求
type CancelOrderReq struct {
	OrderID string `json:"orderId"`
	Reason  string `json:"reason"`
}

// CancelOrderResp 取消订单响应
type CancelOrderResp struct {
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// PayOrderReq 支付订单请求
type PayOrderReq struct {
	OrderID       string `json:"orderId"`
	PaymentMethod int    `json:"paymentMethod"`
	TransactionID string `json:"transactionId"`
}

// PayOrderResp 支付订单响应
type PayOrderResp struct {
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// OrderDTO 订单DTO
type OrderDTO struct {
	OrderID        string           `json:"orderId"`
	PreOrderID     string           `json:"preOrderId"`
	UserID         int64            `json:"userId"`
	CouponID       string           `json:"couponId"`
	OriginalAmount int64            `json:"originalAmount"`
	DiscountAmount int64            `json:"discountAmount"`
	PayableAmount  int64            `json:"payableAmount"`
	PaidAmount     int64            `json:"paidAmount"`
	OrderStatus    int64            `json:"orderStatus"`
	PaymentStatus  int64            `json:"paymentStatus"`
	PaymentMethod  int              `json:"paymentMethod"`
	TransactionID  string           `json:"transactionId"`
	PaidAt         *int64           `json:"paidAt,omitempty"`
	ExpireTime     int64            `json:"expireTime"`
	Items          []*OrderItemDTO  `json:"items"`
	Address        *OrderAddressDTO `json:"address"`
	CreatedAt      int64            `json:"createdAt"`
}

// OrderItemDTO 订单项DTO
type OrderItemDTO struct {
	ProductID   int64  `json:"productId"`
	ProductName string `json:"productName"`
	ProductDesc string `json:"productDesc"`
	Quantity    int    `json:"quantity"`
	Price       int64  `json:"price"`
	TotalPrice  int64  `json:"totalPrice"`
}

// OrderAddressDTO 订单地址DTO
type OrderAddressDTO struct {
	AddressID       int64  `json:"addressId"`
	RecipientName   string `json:"recipientName"`
	PhoneNumber     string `json:"phoneNumber"`
	Province        string `json:"province"`
	City            string `json:"city"`
	DetailedAddress string `json:"detailedAddress"`
}

// OrderListItemDTO 订单列表项DTO
type OrderListItemDTO struct {
	OrderID       string `json:"orderId"`
	TotalAmount   int64  `json:"totalAmount"`
	OrderStatus   int64  `json:"orderStatus"`
	PaymentStatus int64  `json:"paymentStatus"`
	ItemCount     int    `json:"itemCount"`
	CreatedAt     int64  `json:"createdAt"`
}

// GetOrder2PaymentReq 获取订单支付信息请求
type GetOrder2PaymentReq struct {
	OrderID string `json:"orderId"`
}

// GetOrder2PaymentResp 获取订单支付信息响应
type GetOrder2PaymentResp struct {
	OrderID       string `json:"orderId"`
	PayableAmount int64  `json:"payableAmount"`
	StatusCode    int64  `json:"statusCode"`
	StatusMsg     string `json:"statusMsg"`
}
