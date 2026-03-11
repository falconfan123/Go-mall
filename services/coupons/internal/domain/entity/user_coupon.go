package entity

import (
	"errors"
	"time"
)

// UserCouponStatus 用户优惠券状态
type UserCouponStatus int64

const (
	UserCouponStatusUnused  UserCouponStatus = 1 // 未使用
	UserCouponStatusUsed    UserCouponStatus = 2 // 已使用
	UserCouponStatusExpired UserCouponStatus = 3 // 已过期
)

var (
	ErrInvalidUserCouponStatus = errors.New("invalid user coupon status")
	ErrUserCouponAlreadyUsed   = errors.New("user coupon already used")
	ErrUserCouponExpired       = errors.New("user coupon has expired")
)

// UserCoupon 用户优惠券实体
type UserCoupon struct {
	ID       int64            // 记录ID
	UserID   int64            // 用户ID
	CouponID string           // 优惠券ID
	Status   UserCouponStatus // 状态
	GetTime  time.Time        // 领取时间
	UseTime  *time.Time       // 使用时间
	OrderID  *string          // 关联订单ID
}

// NewUserCoupon 创建用户优惠券
func NewUserCoupon(userID int64, couponID string) *UserCoupon {
	return &UserCoupon{
		UserID:   userID,
		CouponID: couponID,
		Status:   UserCouponStatusUnused,
		GetTime:  time.Now(),
	}
}

// Use 使用优惠券
func (uc *UserCoupon) Use(orderID string) error {
	if uc.Status == UserCouponStatusUsed {
		return ErrUserCouponAlreadyUsed
	}
	if uc.Status == UserCouponStatusExpired {
		return ErrUserCouponExpired
	}

	now := time.Now()
	uc.Status = UserCouponStatusUsed
	uc.UseTime = &now
	uc.OrderID = &orderID
	return nil
}

// Expire 过期优惠券
func (uc *UserCoupon) Expire() {
	if uc.Status == UserCouponStatusUnused {
		uc.Status = UserCouponStatusExpired
	}
}

// CancelUse 取消使用
func (uc *UserCoupon) CancelUse() error {
	if uc.Status != UserCouponStatusUsed {
		return errors.New("coupon is not used")
	}

	uc.Status = UserCouponStatusUnused
	uc.UseTime = nil
	uc.OrderID = nil
	return nil
}

// CanUse 是否可使用
func (uc *UserCoupon) CanUse() error {
	if uc.Status != UserCouponStatusUnused {
		switch uc.Status {
		case UserCouponStatusUsed:
			return ErrUserCouponAlreadyUsed
		case UserCouponStatusExpired:
			return ErrUserCouponExpired
		default:
			return ErrInvalidUserCouponStatus
		}
	}
	return nil
}

// IsUnused 是否未使用
func (uc *UserCoupon) IsUnused() bool {
	return uc.Status == UserCouponStatusUnused
}

// IsUsed 是否已使用
func (uc *UserCoupon) IsUsed() bool {
	return uc.Status == UserCouponStatusUsed
}

// IsExpired 是否已过期
func (uc *UserCoupon) IsExpired() bool {
	return uc.Status == UserCouponStatusExpired
}
