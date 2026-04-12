package repository

import (
	"context"

	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/entity"
)

// CheckoutRepository 预订单仓储接口
type CheckoutRepository interface {
	// GetByID 根据预订单ID查询
	GetByID(ctx context.Context, preOrderID string) (*entity.Checkout, error)

	// GetByUserID 根据用户ID查询
	GetByUserID(ctx context.Context, userID int64) ([]*entity.Checkout, error)

	// Save 保存预订单
	Save(ctx context.Context, checkout *entity.Checkout) error

	// Update 更新预订单
	Update(ctx context.Context, checkout *entity.Checkout) error

	// Delete 删除预订单
	Delete(ctx context.Context, preOrderID string) error

	// ListByUserID 查询用户的预订单列表
	ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*entity.Checkout, int64, error)

	// FindExpired 查找已过期的预订单
	FindExpired(ctx context.Context, limit int) ([]*entity.Checkout, error)

	// DecreaseStock 原子扣减库存
	DecreaseStock(ctx context.Context, items []*entity.CheckoutItem) error

	// IncreaseStock 原子恢复库存
	IncreaseStock(ctx context.Context, items []*entity.CheckoutItem) error
}
