package aggregate

import (
	"errors"
	"time"

	"github.com/falconfan123/Go-mall/services/carts/internal/domain/entity"
)

// Cart 购物车聚合根
type Cart struct {
	UserID    int64              // 用户ID
	Items     []*entity.CartItem // 购物车项列表
	UpdatedAt time.Time          // 更新时间
}

var (
	ErrItemNotFound = errors.New("cart item not found")
	ErrItemExists   = errors.New("cart item already exists")
)

// NewCart 创建新购物车
func NewCart(userID int64) *Cart {
	return &Cart{
		UserID:    userID,
		Items:     make([]*entity.CartItem, 0),
		UpdatedAt: time.Now(),
	}
}

// AddItem 添加商品到购物车
func (c *Cart) AddItem(item *entity.CartItem) error {
	// 检查商品是否已存在
	for _, existingItem := range c.Items {
		if existingItem.ProductID == item.ProductID {
			// 已存在，数量+1
			return existingItem.IncreaseQuantity(1)
		}
	}

	// 不存在，添加新项
	c.Items = append(c.Items, item)
	c.UpdatedAt = time.Now()
	return nil
}

// IncreaseItemQuantity 增加购物车项数量
func (c *Cart) IncreaseItemQuantity(productID int64, delta int32) error {
	item := c.findItemByProductID(productID)
	if item == nil {
		return ErrItemNotFound
	}

	if err := item.IncreaseQuantity(delta); err != nil {
		return err
	}

	c.UpdatedAt = time.Now()
	return nil
}

// DecreaseItemQuantity 减少购物车项数量
func (c *Cart) DecreaseItemQuantity(productID int64, delta int32) error {
	item := c.findItemByProductID(productID)
	if item == nil {
		return ErrItemNotFound
	}

	if err := item.DecreaseQuantity(delta); err != nil {
		return err
	}

	// 如果数量为0，删除该项
	if item.Quantity.Value() == 0 {
		c.removeItemByProductID(productID)
	}

	c.UpdatedAt = time.Now()
	return nil
}

// RemoveItem 删除购物车项
func (c *Cart) RemoveItem(productID int64) error {
	if !c.removeItemByProductID(productID) {
		return ErrItemNotFound
	}

	c.UpdatedAt = time.Now()
	return nil
}

// ToggleItemCheck 切换购物车项选中状态
func (c *Cart) ToggleItemCheck(productID int64) error {
	item := c.findItemByProductID(productID)
	if item == nil {
		return ErrItemNotFound
	}

	item.ToggleCheck()
	c.UpdatedAt = time.Now()
	return nil
}

// CheckAll 全选
func (c *Cart) CheckAll() {
	for _, item := range c.Items {
		item.SetChecked(true)
	}
	c.UpdatedAt = time.Now()
}

// UncheckAll 取消全选
func (c *Cart) UncheckAll() {
	for _, item := range c.Items {
		item.SetChecked(false)
	}
	c.UpdatedAt = time.Now()
}

// Clear 清空购物车
func (c *Cart) Clear() {
	c.Items = make([]*entity.CartItem, 0)
	c.UpdatedAt = time.Now()
}

// GetCheckedItems 获取已选中的商品
func (c *Cart) GetCheckedItems() []*entity.CartItem {
	checkedItems := make([]*entity.CartItem, 0)
	for _, item := range c.Items {
		if item.Checked {
			checkedItems = append(checkedItems, item)
		}
	}
	return checkedItems
}

// GetTotalQuantity 获取购物车总商品数量
func (c *Cart) GetTotalQuantity() int32 {
	var total int32 = 0
	for _, item := range c.Items {
		total += item.Quantity.Value()
	}
	return total
}

// GetTotalAmount 获取购物车总金额
func (c *Cart) GetTotalAmount() float64 {
	var total float64 = 0
	for _, item := range c.Items {
		if item.Checked {
			total += item.ProductPrice * float64(item.Quantity.Value())
		}
	}
	return total
}

// GetItemQuantity 获取商品数量
func (c *Cart) GetItemQuantity(productID int64) (int32, error) {
	item := c.findItemByProductID(productID)
	if item == nil {
		return 0, ErrItemNotFound
	}
	return item.Quantity.Value(), nil
}

// findItemByProductID 根据商品ID查找购物车项
func (c *Cart) findItemByProductID(productID int64) *entity.CartItem {
	for _, item := range c.Items {
		if item.ProductID == productID {
			return item
		}
	}
	return nil
}

// removeItemByProductID 根据商品ID删除购物车项
func (c *Cart) removeItemByProductID(productID int64) bool {
	for idx, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:idx], c.Items[idx+1:]...)
			return true
		}
	}
	return false
}
