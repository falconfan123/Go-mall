package dto

// ProductDTO 商品DTO，与protobuf消息结构一致
type ProductDTO struct {
	ID           uint32   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Picture      string   `json:"picture"`
	Price        int64    `json:"price"` // 单位：分
	Stock        int64    `json:"stock"`
	Sold         int64    `json:"sold"`
	Categories   []string `json:"categories"`
	CreatedAt    string   `json:"crated_at"` // 注意：与protobuf字段名保持一致（有拼写错误但兼容）
	UpdatedAt    string   `json:"updated_at"`
	ThumbnailURL string   `json:"thumbnail_url"`
}

// CreateProductRequest 创建商品请求DTO
type CreateProductRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Picture     []byte   `json:"picture"`
	Price       int64    `json:"price"`
	Stock       int64    `json:"stock"`
	Categories  []string `json:"categories"`
}

// CreateProductResponse 创建商品响应DTO
type CreateProductResponse struct {
	StatusCode uint32 `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	ProductID  int64  `json:"product_id"`
}

// GetProductRequest 获取商品请求DTO
type GetProductRequest struct {
	ID     uint32 `json:"id"`
	UserID int32  `json:"user_id"`
}

// GetProductResponse 获取商品响应DTO
type GetProductResponse struct {
	StatusCode uint32      `json:"status_code"`
	StatusMsg  string      `json:"status_msg"`
	Product    *ProductDTO `json:"product"`
}

// UpdateProductRequest 更新商品请求DTO
type UpdateProductRequest struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Picture     []byte   `json:"picture"`
	Price       int64    `json:"price"`
	Stock       int64    `json:"stock"`
	Categories  []string `json:"categories"`
}

// UpdateProductResponse 更新商品响应DTO
type UpdateProductResponse struct {
	StatusCode uint32 `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	ID         int64  `json:"id"`
}

// DeleteProductRequest 删除商品请求DTO
type DeleteProductRequest struct {
	ID int64 `json:"id"`
}

// DeleteProductResponse 删除商品响应DTO
type DeleteProductResponse struct {
	StatusCode uint32 `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

// ListProductsRequest 游标分页查询商品请求DTO
type ListProductsRequest struct {
	Cursor int64 `json:"cursor"`
	Limit  int64 `json:"limit"`
}

// ListProductsResponse 游标分页查询商品响应DTO
type ListProductsResponse struct {
	StatusCode uint32        `json:"status_code"`
	StatusMsg  string        `json:"status_msg"`
	Products   []*ProductDTO `json:"products"`
	NextCursor int64         `json:"nextCursor"`
	HasMore    bool          `json:"hasMore"`
}

// GetAllProductsRequest 分页查询所有商品请求DTO
type GetAllProductsRequest struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
}

// GetAllProductsResponse 分页查询所有商品响应DTO
type GetAllProductsResponse struct {
	StatusCode uint32        `json:"status_code"`
	StatusMsg  string        `json:"status_msg"`
	Total      int64         `json:"total"`
	Products   []*ProductDTO `json:"products"`
	Page       int64         `json:"page"`
	PageSize   int64         `json:"pageSize"`
}

// IsExistProductRequest 检查商品是否存在请求DTO
type IsExistProductRequest struct {
	ID int64 `json:"id"`
}

// IsExistProductResponse 检查商品是否存在响应DTO
type IsExistProductResponse struct {
	StatusCode uint32 `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	Exist      bool   `json:"exist"`
}

// QueryProductRequest 条件查询商品请求DTO
type QueryProductRequest struct {
	Name      string      `json:"name"`
	New       bool        `json:"new"`
	Hot       bool        `json:"hot"`
	Keyword   string      `json:"keyword"`
	Category  []string    `json:"category"`
	Price     *PriceRange `json:"price"`
	Paginator *Paginator  `json:"paginator"`
}

// PriceRange 价格范围DTO
type PriceRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// Paginator 分页参数DTO
type Paginator struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
}

// RecommendProductRequest 推荐商品请求DTO
type RecommendProductRequest struct {
	UserID    int32      `json:"userId"`
	Paginator *Paginator `json:"paginator"`
	Category  []string   `json:"category"`
}

// GetUploadURLRequest 获取上传URL请求DTO
type GetUploadURLRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

// GetUploadURLResponse 获取上传URL响应DTO
type GetUploadURLResponse struct {
	StatusCode uint32            `json:"status_code"`
	StatusMsg  string            `json:"status_msg"`
	UploadURL  string            `json:"uploadUrl"`
	Key        string            `json:"key"`
	FormData   map[string]string `json:"formData"`
}
