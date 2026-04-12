package repository

import (
	"context"

	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/entity"
)

// UserCouponRepository 用户优惠券仓储接口
type UserCouponRepository interface {
	// GetByID 根据ID查询用户优惠券
	GetByID(ctx context.Context, id int64) (*entity.UserCoupon, error)

	// GetByUserIDAndCouponID 根据用户ID和优惠券ID查询
	GetByUserIDAndCouponID(ctx context.Context, userID int64, couponID string) (*entity.UserCoupon, error)

	// ListByUserID 查询用户的优惠券列表
	ListByUserID(ctx context.Context, userID int64, status *entity.UserCouponStatus, page, pageSize int) ([]*entity.UserCoupon, int64, error)

	// Save 保存用户优惠券
	Save(ctx context.Context, userCoupon *entity.UserCoupon) error

	// Update 更新用户优惠券
	Update(ctx context.Context, userCoupon *entity.UserCoupon) error

	// Delete 删除用户优惠券
	Delete(ctx context.Context, id int64) error

	// CountByUserIDAndCouponID 统计用户领取某优惠券的数量
	CountByUserIDAndCouponID(ctx context.Context, userID int64, couponID string) (int64, error)

	// FindAvailableByUserID 查询用户可用的优惠券
	FindAvailableByUserID(ctx context.Context, userID int64, orderAmount int64) ([]*entity.UserCoupon, error)
}
