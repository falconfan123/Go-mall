package aggregate

import (
	"errors"
	"time"

	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/valueobject"
)

var (
	ErrCheckoutNotFound    = errors.New("checkout not found")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrCheckoutExpired     = errors.New("checkout expired")
	ErrInvalidCheckoutItem = errors.New("invalid checkout item")
)

// CheckoutAggregate 预订单聚合根
type CheckoutAggregate struct {
	checkout *entity.Checkout
}

// NewCheckoutAggregate 创建预订单聚合根
func NewCheckoutAggregate(
	preOrderID string,
	userID int64,
	addressID int64,
	expireMinutes int,
) *CheckoutAggregate {
	return &CheckoutAggregate{
		checkout: entity.NewCheckout(preOrderID, userID, addressID, 0, expireMinutes),
	}
}

// LoadCheckout 加载预订单到聚合根
func LoadCheckout(checkout *entity.Checkout) *CheckoutAggregate {
	return &CheckoutAggregate{
		checkout: checkout,
	}
}

// GetCheckout 获取预订单
func (c *CheckoutAggregate) GetCheckout() *entity.Checkout {
	return c.checkout
}

// AddItem 添加商品到预订单
func (c *CheckoutAggregate) AddItem(
	productID int64,
	productName string,
	productImage string,
	quantity int,
	price int64,
) error {
	if quantity <= 0 {
		return ErrInvalidCheckoutItem
	}

	snapshot := &entity.ProductSnapshot{
		ProductName:  productName,
		ProductImage: productImage,
	}

	item := entity.NewCheckoutItem(
		c.checkout.PreOrderID,
		productID,
		quantity,
		price,
		snapshot,
	)

	c.checkout.AddItem(item)
	return nil
}

// CalculateTotal 计算总金额
func (c *CheckoutAggregate) CalculateTotal() int64 {
	return c.checkout.CalculateTotal()
}

// ApplyDiscount 应用优惠
func (c *CheckoutAggregate) ApplyDiscount(discountAmount int64) error {
	return c.checkout.ApplyDiscount(discountAmount)
}

// Confirm 确认预订单
func (c *CheckoutAggregate) Confirm() error {
	return c.checkout.Confirm()
}

// Cancel 取消预订单
func (c *CheckoutAggregate) Cancel() error {
	return c.checkout.Cancel()
}

// Expire 过期预订单
func (c *CheckoutAggregate) Expire() {
	c.checkout.Expire()
}

// IsExpired 检查是否过期
func (c *CheckoutAggregate) IsExpired() bool {
	return c.checkout.IsExpired()
}

// IsPending 检查是否处于预占状态
func (c *CheckoutAggregate) IsPending() bool {
	return c.checkout.IsPending()
}

// GetItems 获取商品列表
func (c *CheckoutAggregate) GetItems() []*entity.CheckoutItem {
	return c.checkout.Items
}

// GetTotalQuantity 获取商品总数量
func (c *CheckoutAggregate) GetTotalQuantity() int {
	return c.checkout.GetTotalQuantity()
}

// GetFinalAmount 获取实付金额
func (c *CheckoutAggregate) GetFinalAmount() int64 {
	return c.checkout.FinalAmount
}

// GetOriginalAmount 获取原始金额
func (c *CheckoutAggregate) GetOriginalAmount() int64 {
	return c.checkout.OriginalAmount
}

// GetUserID 获取用户ID
func (c *CheckoutAggregate) GetUserID() int64 {
	return c.checkout.UserID
}

// GetAddressID 获取地址ID
func (c *CheckoutAggregate) GetAddressID() int64 {
	return c.checkout.AddressID
}

// GetPreOrderID 获取预订单ID
func (c *CheckoutAggregate) GetPreOrderID() string {
	return c.checkout.PreOrderID
}

// GetExpireTime 获取过期时间
func (c *CheckoutAggregate) GetExpireTime() time.Time {
	return c.checkout.ExpireTime
}

// CreateFromCart 从购物车创建预订单
func (c *CheckoutAggregate) CreateFromCart(
	cartItems []*CartItemInfo,
	address *valueobject.Address,
	expireMinutes int,
) error {
	// 验证地址
	if address == nil || address.ID == 0 {
		return errors.New("invalid address")
	}

	c.checkout.AddressID = address.ID
	c.checkout.UserID = address.UserID

	// 添加商品
	var totalAmount int64
	for _, item := range cartItems {
		if err := c.AddItem(
			item.ProductID,
			item.ProductName,
			item.ProductImage,
			item.Quantity,
			item.Price,
		); err != nil {
			return err
		}
		totalAmount += item.Price * int64(item.Quantity)
	}

	c.checkout.OriginalAmount = totalAmount
	c.checkout.FinalAmount = totalAmount

	return nil
}

// CartItemInfo 购物车商品信息
type CartItemInfo struct {
	ProductID    int64
	ProductName  string
	ProductImage string
	Quantity     int
	Price        int64
}
