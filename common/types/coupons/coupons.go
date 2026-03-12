package coupons

// CouponStatus 优惠券状态枚举
type CouponStatus int32

const (
	CouponStatusUnspecified CouponStatus = 0
	CouponStatusAvailable   CouponStatus = 1 // 可用
	CouponStatusLocked      CouponStatus = 2 // 已锁定（下单时使用）
	CouponStatusUsed        CouponStatus = 3 // 已使用
	CouponStatusExpired     CouponStatus = 4 // 已过期
)
