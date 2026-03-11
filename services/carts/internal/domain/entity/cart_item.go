package entity

import (
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/valueobject"
)

// CartItem 购物车项实体
type CartItem struct {
	ID          int64                   // 购物车项ID
	ProductID   int64                   // 商品ID
	ProductName string                  // 商品名称
	ProductImage string                 // 商品图片
	ProductPrice float64                // 商品价格
	Quantity    valueobject.Quantity    // 商品数量
	Checked     bool                    // 是否选中
}

// NewCartItem 创建购物车项
func NewCartItem(
	productID int64,
	productName string,
	productImage string,
	productPrice float64,
	quantity valueobject.Quantity,
) *CartItem {
	return &CartItem{
		ProductID:   productID,
		ProductName: productName,
		ProductImage: productImage,
		ProductPrice: productPrice,
		Quantity:    quantity,
		Checked:     true, // 默认选中
	}
}

// IncreaseQuantity 增加数量
func (i *CartItem) IncreaseQuantity(delta int32) error {
	newQty, err := i.Quantity.Add(delta)
	if err != nil {
		return err
	}
	i.Quantity = newQty
	return nil
}

// DecreaseQuantity 减少数量
func (i *CartItem) DecreaseQuantity(delta int32) error {
	newQty, err := i.Quantity.Subtract(delta)
	if err != nil {
		return err
	}
	i.Quantity = newQty
	return nil
}

// ToggleCheck 切换选中状态
func (i *CartItem) ToggleCheck() {
	i.Checked = !i.Checked
}

// SetChecked 设置选中状态
func (i *CartItem) SetChecked(checked bool) {
	i.Checked = checked
}
