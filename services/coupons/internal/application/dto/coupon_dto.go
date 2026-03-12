package dto

// CreateCouponReq 创建优惠券请求
type CreateCouponReq struct {
	Name       string `json:"name"`
	CouponType int64  `json:"couponType"` // 1: 满减券, 2: 折扣券
	Value      int64  `json:"value"`      // 优惠值(满减:金额(分), 折扣:折扣值(1-100))
	MinAmount  int64  `json:"minAmount"`  // 最低消费金额(分)
	TotalCount uint64 `json:"totalCount"` // 发行总量
	StartTime  int64  `json:"startTime"`  // 开始时间戳
	EndTime    int64  `json:"endTime"`    // 结束时间戳
}

// CreateCouponResp 创建优惠券响应
type CreateCouponResp struct {
	CouponID   string `json:"couponId"`
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// ClaimCouponReq 领取优惠券请求
type ClaimCouponReq struct {
	UserID   int64  `json:"userId"`
	CouponID string `json:"couponId"`
}

// ClaimCouponResp 领取优惠券响应
type ClaimCouponResp struct {
	Coupon     *CouponDTO `json:"coupon"`
	StatusCode int64      `json:"statusCode"`
	StatusMsg  string     `json:"statusMsg"`
}

// UseCouponReq 使用优惠券请求
type UseCouponReq struct {
	UserID   int64  `json:"userId"`
	CouponID int64  `json:"couponId"`
	OrderID  string `json:"orderId"`
}

// UseCouponResp 使用优惠券响应
type UseCouponResp struct {
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// CalculateCouponReq 计算优惠金额请求
type CalculateCouponReq struct {
	UserID      int64  `json:"userId"`
	CouponID    int64  `json:"couponId"`
	OrderID     string `json:"orderId"`
	OrderAmount int64  `json:"orderAmount"`
}

// CalculateCouponResp 计算优惠金额响应
type CalculateCouponResp struct {
	DiscountAmount int64  `json:"discountAmount"`
	FinalAmount    int64  `json:"finalAmount"`
	StatusCode     int64  `json:"statusCode"`
	StatusMsg      string `json:"statusMsg"`
}

// ListCouponsReq 查询优惠券列表请求
type ListCouponsReq struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// ListCouponsResp 查询优惠券列表响应
type ListCouponsResp struct {
	Coupons    []*CouponDTO `json:"coupons"`
	TotalCount int64        `json:"totalCount"`
	StatusCode int64        `json:"statusCode"`
	StatusMsg  string       `json:"statusMsg"`
}

// ListUserCouponsReq 查询用户优惠券请求
type ListUserCouponsReq struct {
	UserID   int64 `json:"userId"`
	Status   *int  `json:"status"` // 1:未使用, 2:已使用, 3:已过期, nil:全部
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// ListUserCouponsResp 查询用户优惠券响应
type ListUserCouponsResp struct {
	UserCoupons []*UserCouponDTO `json:"userCoupons"`
	TotalCount  int64            `json:"totalCount"`
	StatusCode  int64            `json:"statusCode"`
	StatusMsg   string           `json:"statusMsg"`
}

// CouponDTO 优惠券DTO
type CouponDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CouponType     int64  `json:"couponType"`
	Value          int64  `json:"value"`
	MinAmount      int64  `json:"minAmount"`
	TotalCount     uint64 `json:"totalCount"`
	RemainingCount uint64 `json:"remainingCount"`
	StartTime      int64  `json:"startTime"`
	EndTime        int64  `json:"endTime"`
	Status         int64  `json:"status"`
}

// UserCouponDTO 用户优惠券DTO
type UserCouponDTO struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"userId"`
	CouponID string `json:"couponId"`
	Status   int64  `json:"status"`
	GetTime  int64  `json:"getTime"`
	UseTime  *int64 `json:"useTime,omitempty"`
	OrderID  string `json:"orderId,omitempty"`
	// 优惠券详情
	Coupon *CouponDTO `json:"coupon,omitempty"`
}

// LockCouponReq 锁定优惠券请求
type LockCouponReq struct {
	UserID   int64  `json:"userId"`
	CouponID int64  `json:"couponId"`
	OrderID  string `json:"orderId"`
}

// LockCouponResp 锁定优惠券响应
type LockCouponResp struct {
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// ReleaseCouponReq 释放优惠券请求
type ReleaseCouponReq struct {
	UserID   int64  `json:"userId"`
	CouponID int64  `json:"couponId"`
	OrderID  string `json:"orderId"`
}

// ReleaseCouponResp 释放优惠券响应
type ReleaseCouponResp struct {
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// GetCouponReq 获取优惠券详情请求
type GetCouponReq struct {
	CouponID string `json:"couponId"`
}

// GetCouponResp 获取优惠券详情响应
type GetCouponResp struct {
	Coupon     *CouponDTO `json:"coupon"`
	StatusCode int64      `json:"statusCode"`
	StatusMsg  string     `json:"statusMsg"`
}
