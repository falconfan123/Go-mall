package entity

import (
	"errors"
	"time"
)

// CheckoutStatus 预订单状态
type CheckoutStatus int

const (
	CheckoutStatusPending   CheckoutStatus = 0 // 预占中
	CheckoutStatusConfirmed CheckoutStatus = 1 // 已确认
	CheckoutStatusCanceled  CheckoutStatus = 2 // 已取消
	CheckoutStatusExpired   CheckoutStatus = 3 // 已过期
)

var (
	ErrInvalidCheckoutStatus = errors.New("invalid checkout status")
	ErrCheckoutExpired       = errors.New("checkout has expired")
	ErrCheckoutAlreadyUsed   = errors.New("checkout already used")
)

// Checkout 预订单实体
type Checkout struct {
	PreOrderID     string          // 预订单ID
	UserID         int64           // 用户ID
	AddressID      int64           // 收货地址ID
	CouponIDs      []string        // 优惠券ID列表
	OriginalAmount int64           // 原始金额（分）
	FinalAmount    int64           // 实付金额（分）
	Status         CheckoutStatus  // 状态
	ExpireTime     time.Time       // 过期时间
	Items          []*CheckoutItem // 商品明细
	CreatedAt      time.Time       // 创建时间
	UpdatedAt      time.Time       // 更新时间
}

// NewCheckout 创建预订单
func NewCheckout(
	preOrderID string,
	userID int64,
	addressID int64,
	originalAmount int64,
	expireMinutes int,
) *Checkout {
	return &Checkout{
		PreOrderID:     preOrderID,
		UserID:         userID,
		AddressID:      addressID,
		OriginalAmount: originalAmount,
		FinalAmount:    originalAmount,
		Status:         CheckoutStatusPending,
		ExpireTime:     time.Now().Add(time.Duration(expireMinutes) * time.Minute),
		Items:          make([]*CheckoutItem, 0),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// AddItem 添加商品
func (c *Checkout) AddItem(item *CheckoutItem) {
	c.Items = append(c.Items, item)
}

// CalculateTotal 计算总金额
func (c *Checkout) CalculateTotal() int64 {
	var total int64
	for _, item := range c.Items {
		total += item.TotalPrice()
	}
	c.OriginalAmount = total
	c.FinalAmount = total
	return total
}

// ApplyDiscount 应用优惠
func (c *Checkout) ApplyDiscount(discountAmount int64) error {
	if discountAmount < 0 {
		return errors.New("discount amount cannot be negative")
	}
	if discountAmount > c.OriginalAmount {
		return errors.New("discount amount cannot exceed original amount")
	}
	c.FinalAmount = c.OriginalAmount - discountAmount
	return nil
}

// Confirm 确认预订单
func (c *Checkout) Confirm() error {
	if c.Status != CheckoutStatusPending {
		return ErrInvalidCheckoutStatus
	}
	if time.Now().After(c.ExpireTime) {
		c.Status = CheckoutStatusExpired
		return ErrCheckoutExpired
	}
	c.Status = CheckoutStatusConfirmed
	c.UpdatedAt = time.Now()
	return nil
}

// Cancel 取消预订单
func (c *Checkout) Cancel() error {
	if c.Status == CheckoutStatusConfirmed {
		return ErrCheckoutAlreadyUsed
	}
	c.Status = CheckoutStatusCanceled
	c.UpdatedAt = time.Now()
	return nil
}

// Expire 过期预订单
func (c *Checkout) Expire() {
	if c.Status == CheckoutStatusPending {
		c.Status = CheckoutStatusExpired
		c.UpdatedAt = time.Now()
	}
}

// IsExpired 是否已过期
func (c *Checkout) IsExpired() bool {
	return time.Now().After(c.ExpireTime)
}

// IsPending 是否处于预占状态
func (c *Checkout) IsPending() bool {
	return c.Status == CheckoutStatusPending
}

// GetTotalQuantity 获取商品总数量
func (c *Checkout) GetTotalQuantity() int {
	var total int
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}

// CheckoutItem 预订单商品项实体
type CheckoutItem struct {
	ID         int64            // 项ID
	PreOrderID string           // 预订单ID
	ProductID  int64            // 商品ID
	Quantity   int              // 数量
	Price      int64            // 单价（分）
	Snapshot   *ProductSnapshot // 商品快照
	CreatedAt  time.Time        // 创建时间
}

// ProductSnapshot 商品快照
type ProductSnapshot struct {
	ProductName  string `json:"productName"`  // 商品名称
	ProductImage string `json:"productImage"` // 商品图片
	ProductSpecs string `json:"productSpecs"` // 规格属性
}

// NewCheckoutItem 创建预订单商品项
func NewCheckoutItem(
	preOrderID string,
	productID int64,
	quantity int,
	price int64,
	snapshot *ProductSnapshot,
) *CheckoutItem {
	return &CheckoutItem{
		PreOrderID: preOrderID,
		ProductID:  productID,
		Quantity:   quantity,
		Price:      price,
		Snapshot:   snapshot,
		CreatedAt:  time.Now(),
	}
}

// TotalPrice 计算商品总价
func (i *CheckoutItem) TotalPrice() int64 {
	return i.Price * int64(i.Quantity)
}
