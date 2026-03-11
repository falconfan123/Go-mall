package repository

import (
	"context"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/aggregate"
)

// CartRepository 购物车仓储接口
type CartRepository interface {
	// GetByUserID 根据用户ID查询购物车
	GetByUserID(ctx context.Context, userID int64) (*aggregate.Cart, error)

	// Save 保存购物车
	Save(ctx context.Context, cart *aggregate.Cart) error

	// AddItem 添加购物车项
	AddItem(ctx context.Context, userID int64, item *aggregate.Cart) error

	// UpdateItemQuantity 更新购物车项数量
	UpdateItemQuantity(ctx context.Context, userID int64, productID int64, quantity int32) error

	// RemoveItem 删除购物车项
	RemoveItem(ctx context.Context, userID int64, productID int64) error

	// Clear 清空购物车
	Clear(ctx context.Context, userID int64) error
}
