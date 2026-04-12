package entity

import (
	"time"
)

// ProductCategory 分类实体
type ProductCategory struct {
	ID   int64  // 分类ID
	Name string // 分类名称
}

// ProductImage 商品图片实体
type ProductImage struct {
	ID        int64     // 图片ID
	ProductID int64     // 商品ID
	URL       string    // 图片URL
	IsPrimary bool      // 是否主图
	SortOrder int       // 排序
	CreatedAt time.Time // 创建时间
}

// NewProductImage 创建商品图片
func NewProductImage(productID int64, url string, isPrimary bool, sortOrder int) *ProductImage {
	return &ProductImage{
		ProductID: productID,
		URL:       url,
		IsPrimary: isPrimary,
		SortOrder: sortOrder,
		CreatedAt: time.Now(),
	}
}

// ProductSpec 商品规格实体
type ProductSpec struct {
	ID         int64     // 规格ID
	ProductID  int64     // 商品ID
	SpecName   string    // 规格名称（如"颜色"、"尺寸"）
	SpecValue  string    // 规格值（如"红色"、"XL"）
	Stock      int64     // 规格库存
	PriceDelta int64     // 价格增量（分）
	CreatedAt  time.Time // 创建时间
}

// ProductReview 商品评价实体
type ProductReview struct {
	ID          int64        // 评价ID
	ProductID   int64        // 商品ID
	UserID      int64        // 用户ID
	Rating      int          // 评分（1-5）
	Content     string       // 评价内容
	Images      []string     // 评价图片
	IsAnonymous bool         // 是否匿名
	Status      ReviewStatus // 状态
	CreatedAt   time.Time    // 创建时间
	UpdatedAt   time.Time    // 更新时间
}

// ReviewStatus 评价状态
type ReviewStatus int

const (
	ReviewStatusPending  ReviewStatus = 1 // 待审核
	ReviewStatusApproved ReviewStatus = 2 // 已通过
	ReviewStatusRejected ReviewStatus = 3 // 已拒绝
)

// NewProductReview 创建商品评价
func NewProductReview(productID, userID int64, rating int, content string, isAnonymous bool) *ProductReview {
	return &ProductReview{
		ProductID:   productID,
		UserID:      userID,
		Rating:      rating,
		Content:     content,
		IsAnonymous: isAnonymous,
		Status:      ReviewStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Approve 审核通过
func (r *ProductReview) Approve() {
	r.Status = ReviewStatusApproved
	r.UpdatedAt = time.Now()
}

// Reject 审核拒绝
func (r *ProductReview) Reject() {
	r.Status = ReviewStatusRejected
	r.UpdatedAt = time.Now()
}
