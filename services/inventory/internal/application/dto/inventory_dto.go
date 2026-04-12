package dto

// InventoryItem 库存项DTO
type InventoryItem struct {
	ProductID int32 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

// PreDecreaseInventoryRequest 预扣库存请求DTO
type PreDecreaseInventoryRequest struct {
	Items      []*InventoryItem `json:"items"`
	PreOrderID string           `json:"pre_order_id"`
	UserID     int32            `json:"user_id"`
}

// DecreaseInventoryRequest 扣减库存请求DTO
type DecreaseInventoryRequest struct {
	Items      []*InventoryItem `json:"items"`
	PreOrderID string           `json:"pre_order_id"`
	UserID     int32            `json:"user_id"`
}

// ReturnPreInventoryRequest 退还预扣库存请求DTO
type ReturnPreInventoryRequest struct {
	Items      []*InventoryItem `json:"items"`
	PreOrderID string           `json:"pre_order_id"`
	UserID     int32            `json:"user_id"`
}

// ReturnInventoryRequest 退还库存请求DTO
type ReturnInventoryRequest struct {
	Items   []*InventoryItem `json:"items"`
	OrderID string           `json:"order_id"`
	UserID  int32            `json:"user_id"`
}

// UpdateInventoryRequest 更新库存请求DTO
type UpdateInventoryRequest struct {
	Items []*InventoryItem `json:"items"`
}

// GetInventoryRequest 查询库存请求DTO
type GetInventoryRequest struct {
	ProductID int32 `json:"product_id"`
}

// GetInventoryResponse 查询库存响应DTO
type GetInventoryResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	Inventory  int64  `json:"inventory"`
	SoldCount  int64  `json:"sold_count"`
}

// InventoryResponse 通用库存操作响应DTO
type InventoryResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}
