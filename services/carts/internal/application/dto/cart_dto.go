package dto

// AddItemReq 添加商品到购物车请求
type AddItemReq struct {
	UserID       int64   `json:"userId"`
	ProductID    int64   `json:"productId"`
	ProductName  string  `json:"productName"`
	ProductImage string  `json:"productImage"`
	ProductPrice float64 `json:"productPrice"`
	Quantity     int32   `json:"quantity"`
}

// UpdateItemQuantityReq 更新购物车商品数量请求
type UpdateItemQuantityReq struct {
	UserID    int64 `json:"userId"`
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

// RemoveItemReq 删除购物车商品请求
type RemoveItemReq struct {
	UserID    int64 `json:"userId"`
	ProductID int64 `json:"productId"`
}

// ToggleItemCheckReq 切换购物车商品选中状态请求
type ToggleItemCheckReq struct {
	UserID    int64 `json:"userId"`
	ProductID int64 `json:"productId"`
}

// CheckAllReq 全选购物车商品请求
type CheckAllReq struct {
	UserID int64 `json:"userId"`
}

// UncheckAllReq 取消全选购物车商品请求
type UncheckAllReq struct {
	UserID int64 `json:"userId"`
}

// ClearReq 清空购物车请求
type ClearReq struct {
	UserID int64 `json:"userId"`
}

// GetCartReq 查询购物车请求
type GetCartReq struct {
	UserID int64 `json:"userId"`
}

// CartItemDTO 购物车项响应
type CartItemDTO struct {
	ProductID    int64   `json:"productId"`
	ProductName  string  `json:"productName"`
	ProductImage string  `json:"productImage"`
	ProductPrice float64 `json:"productPrice"`
	Quantity     int32   `json:"quantity"`
	Checked      bool    `json:"checked"`
}

// CartDTO 购物车响应
type CartDTO struct {
	UserID        int64          `json:"userId"`
	Items         []*CartItemDTO `json:"items"`
	TotalQuantity int32          `json:"totalQuantity"`
	TotalAmount   float64        `json:"totalAmount"`
	CheckedCount  int32          `json:"checkedCount"`
	CheckedAmount float64        `json:"checkedAmount"`
}
