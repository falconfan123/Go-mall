package repository

import (
	"context"

	"github.com/falconfan123/Go-mall/services/order/internal/domain/entity"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	// GetByID 根据订单ID查询
	GetByID(ctx context.Context, orderID string) (*entity.Order, error)

	// GetByPreOrderID 根据预订单ID查询
	GetByPreOrderID(ctx context.Context, preOrderID string) (*entity.Order, error)

	// GetByUserID 根据用户ID查询
	GetByUserID(ctx context.Context, userID int64) ([]*entity.Order, error)

	// Save 保存订单
	Save(ctx context.Context, order *entity.Order) error

	// Update 更新订单
	Update(ctx context.Context, order *entity.Order) error

	// Delete 删除订单
	Delete(ctx context.Context, orderID string) error

	// ListByUserID 查询用户的订单列表
	ListByUserID(ctx context.Context, userID int64, status *entity.OrderStatus, page, pageSize int) ([]*entity.Order, int64, error)

	// ListByStatus 根据状态查询订单列表
	ListByStatus(ctx context.Context, status entity.OrderStatus, page, pageSize int) ([]*entity.Order, int64, error)

	// FindExpired 查找已过期的订单
	FindExpired(ctx context.Context, limit int) ([]*entity.Order, error)
}
