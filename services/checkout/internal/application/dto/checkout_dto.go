package dto

// PrepareCheckoutReq 准备结算请求
type PrepareCheckoutReq struct {
	UserID    int64              `json:"userId"`
	AddressID int64              `json:"addressId"`
	CouponIDs []string           `json:"couponIds"`
	Items     []*CheckoutItemReq `json:"items"`
}

// CheckoutItemReq 结算商品项请求
type CheckoutItemReq struct {
	ProductID    int64  `json:"productId"`
	ProductName  string `json:"productName"`
	ProductImage string `json:"productImage"`
	Quantity     int    `json:"quantity"`
	Price        int64  `json:"price"`
}

// PrepareCheckoutResp 准备结算响应
type PrepareCheckoutResp struct {
	PreOrderID     string `json:"preOrderId"`
	OriginalAmount int64  `json:"originalAmount"`
	DiscountAmount int64  `json:"discountAmount"`
	FinalAmount    int64  `json:"finalAmount"`
	ExpireTime     int64  `json:"expireTime"`
	StatusCode     int64  `json:"statusCode"`
	StatusMsg      string `json:"statusMsg"`
}

// GetCheckoutDetailReq 获取结算详情请求
type GetCheckoutDetailReq struct {
	PreOrderID string `json:"preOrderId"`
}

// GetCheckoutDetailResp 获取结算详情响应
type GetCheckoutDetailResp struct {
	PreOrderID     string             `json:"preOrderId"`
	UserID         int64              `json:"userId"`
	Address        *AddressDTO        `json:"address"`
	Items          []*CheckoutItemDTO `json:"items"`
	OriginalAmount int64              `json:"originalAmount"`
	DiscountAmount int64              `json:"discountAmount"`
	FinalAmount    int64              `json:"finalAmount"`
	Status         int64              `json:"status"`
	ExpireTime     int64              `json:"expireTime"`
	StatusCode     int64              `json:"statusCode"`
	StatusMsg      string             `json:"statusMsg"`
}

// CheckoutItemDTO 结算商品项DTO
type CheckoutItemDTO struct {
	ProductID    int64  `json:"productId"`
	ProductName  string `json:"productName"`
	ProductImage string `json:"productImage"`
	Quantity     int    `json:"quantity"`
	Price        int64  `json:"price"`
	TotalPrice   int64  `json:"totalPrice"`
}

// AddressDTO 地址DTO
type AddressDTO struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Detail    string `json:"detail"`
	ZipCode   string `json:"zipCode"`
	IsDefault bool   `json:"isDefault"`
}

// ListCheckoutReq 查询结算列表请求
type ListCheckoutReq struct {
	UserID   int64 `json:"userId"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// ListCheckoutResp 查询结算列表响应
type ListCheckoutResp struct {
	Checkouts  []*CheckoutListItemDTO `json:"checkouts"`
	TotalCount int64                  `json:"totalCount"`
	StatusCode int64                  `json:"statusCode"`
	StatusMsg  string                 `json:"statusMsg"`
}

// CheckoutListItemDTO 结算列表项DTO
type CheckoutListItemDTO struct {
	PreOrderID     string `json:"preOrderId"`
	TotalQuantity  int    `json:"totalQuantity"`
	OriginalAmount int64  `json:"originalAmount"`
	FinalAmount    int64  `json:"finalAmount"`
	Status         int64  `json:"status"`
	ExpireTime     int64  `json:"expireTime"`
	CreatedAt      int64  `json:"createdAt"`
}

// ConfirmCheckoutReq 确认结算请求
type ConfirmCheckoutReq struct {
	PreOrderID string `json:"preOrderId"`
}

// ConfirmCheckoutResp 确认结算响应
type ConfirmCheckoutResp struct {
	OrderID    string `json:"orderId"`
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// CancelCheckoutReq 取消结算请求
type CancelCheckoutReq struct {
	PreOrderID string `json:"preOrderId"`
}

// CancelCheckoutResp 取消结算响应
type CancelCheckoutResp struct {
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// ReleaseCheckoutReq 释放结算请求
type ReleaseCheckoutReq struct {
	PreOrderID string `json:"preOrderId"`
}

// ReleaseCheckoutResp 释放结算响应
type ReleaseCheckoutResp struct {
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// UpdateStatusReq 更新状态请求
type UpdateStatusReq struct {
	PreOrderID string `json:"preOrderId"`
	Status     int64  `json:"status"`
}

// UpdateStatusResp 更新状态响应
type UpdateStatusResp struct {
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}
