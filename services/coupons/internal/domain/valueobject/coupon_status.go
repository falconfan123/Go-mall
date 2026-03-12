package valueobject

import "errors"

// CouponStatus 优惠券状态
type CouponStatus int64

const (
	CouponStatusDisabled CouponStatus = 0 // 禁用
	CouponStatusEnabled  CouponStatus = 1 // 启用
)

var (
	ErrInvalidCouponStatus = errors.New("invalid coupon status")
)

// NewCouponStatus 创建优惠券状态
func NewCouponStatus(s int64) (CouponStatus, error) {
	switch s {
	case 0, 1:
		return CouponStatus(s), nil
	default:
		return 0, ErrInvalidCouponStatus
	}
}

// Value 获取状态值
func (s CouponStatus) Value() int64 {
	return int64(s)
}

// IsValid 是否有效
func (s CouponStatus) IsValid() bool {
	return s >= 0 && s <= 1
}

// IsEnabled 是否启用
func (s CouponStatus) IsEnabled() bool {
	return s == CouponStatusEnabled
}

// String 转为字符串
func (s CouponStatus) String() string {
	switch s {
	case CouponStatusDisabled:
		return "禁用"
	case CouponStatusEnabled:
		return "启用"
	default:
		return "未知状态"
	}
}
