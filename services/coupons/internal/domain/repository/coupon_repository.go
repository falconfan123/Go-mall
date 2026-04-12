package repository

import (
	"context"

	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/aggregate"
)

// CouponRepository 优惠券仓储接口
type CouponRepository interface {
	// GetByID 根据ID查询优惠券
	GetByID(ctx context.Context, id string) (*aggregate.Coupon, error)

	// Save 保存优惠券
	Save(ctx context.Context, coupon *aggregate.Coupon) error

	// Update 更新优惠券
	Update(ctx context.Context, coupon *aggregate.Coupon) error

	// Delete 删除优惠券
	Delete(ctx context.Context, id string) error

	// List 查询优惠券列表
	List(ctx context.Context, page, pageSize int) ([]*aggregate.Coupon, int64, error)

	// ListAvailable 查询可用优惠券列表
	ListAvailable(ctx context.Context, page, pageSize int) ([]*aggregate.Coupon, int64, error)

	// DecreaseStock 原子扣减库存
	DecreaseStock(ctx context.Context, couponID string, count int) error

	// IncreaseStock 原子增加库存
	IncreaseStock(ctx context.Context, couponID string, count int) error
}
