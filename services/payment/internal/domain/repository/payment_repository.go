package repository

import (
	"context"

	"github.com/falconfan123/Go-mall/services/payment/internal/domain/entity"
)

// PaymentRepository 支付单仓储接口
type PaymentRepository interface {
	// GetByID 根据支付单ID查询
	GetByID(ctx context.Context, paymentID string) (*entity.Payment, error)

	// GetByOrderID 根据订单ID查询
	GetByOrderID(ctx context.Context, orderID string) (*entity.Payment, error)

	// GetByPreOrderID 根据预订单ID查询
	GetByPreOrderID(ctx context.Context, preOrderID string) (*entity.Payment, error)

	// Save 保存支付单
	Save(ctx context.Context, payment *entity.Payment) error

	// Update 更新支付单
	Update(ctx context.Context, payment *entity.Payment) error

	// Delete 删除支付单
	Delete(ctx context.Context, paymentID string) error

	// ListByUserID 查询用户的支付单列表
	ListByUserID(ctx context.Context, userID int64, status *entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int64, error)

	// ListByStatus 根据状态查询支付单列表
	ListByStatus(ctx context.Context, status entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int64, error)

	// FindExpired 查找已过期的支付单
	FindExpired(ctx context.Context, limit int) ([]*entity.Payment, error)
}
