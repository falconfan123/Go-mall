package valueobject

import "errors"

// CouponType 优惠券类型
type CouponType int64

const (
	CouponTypeFullReduction CouponType = 1 // 满减
	CouponTypeDiscount       CouponType = 2 // 折扣
	CouponTypeDirectReduction CouponType = 3 // 立减
)

var (
	ErrInvalidCouponType = errors.New("invalid coupon type")
)

// NewCouponType 创建优惠券类型
func NewCouponType(t int64) (CouponType, error) {
	switch t {
	case 1, 2, 3:
		return CouponType(t), nil
	default:
		return 0, ErrInvalidCouponType
	}
}

// Value 获取类型值
func (t CouponType) Value() int64 {
	return int64(t)
}

// IsValid 是否有效
func (t CouponType) IsValid() bool {
	return t >= 1 && t <= 3
}

// String 转为字符串
func (t CouponType) String() string {
	switch t {
	case CouponTypeFullReduction:
		return "满减"
	case CouponTypeDiscount:
		return "折扣"
	case CouponTypeDirectReduction:
		return "立减"
	default:
		return "未知类型"
	}
}
