package aggregate

import (
	"errors"
	"time"

	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/valueobject"
)

// Coupon 优惠券聚合根
type Coupon struct {
	ID             string                   // 优惠券ID
	Name           string                   // 券名称
	Discount       *valueobject.Discount    // 折扣信息
	ValidPeriod    *valueobject.ValidPeriod // 有效期
	Status         valueobject.CouponStatus // 状态
	TotalCount     uint64                   // 发行总量
	RemainingCount uint64                   // 剩余数量
	CreatedAt      time.Time                // 创建时间
	UpdatedAt      time.Time                // 更新时间
}

var (
	ErrCouponOutOfStock = errors.New("coupon is out of stock")
	ErrCouponInvalid    = errors.New("invalid coupon")
)

// NewCoupon 创建新优惠券
func NewCoupon(
	id string,
	name string,
	couponType valueobject.CouponType,
	value int64,
	minAmount int64,
	startTime time.Time,
	endTime time.Time,
	totalCount uint64,
) (*Coupon, error) {
	// 创建折扣对象
	discount, err := valueobject.NewDiscount(couponType, value, minAmount)
	if err != nil {
		return nil, err
	}

	// 创建有效期
	validPeriod, err := valueobject.NewValidPeriod(startTime, endTime)
	if err != nil {
		return nil, err
	}

	return &Coupon{
		ID:             id,
		Name:           name,
		Discount:       discount,
		ValidPeriod:    validPeriod,
		Status:         valueobject.CouponStatusEnabled,
		TotalCount:     totalCount,
		RemainingCount: totalCount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

// CanClaim 是否可领取
func (c *Coupon) CanClaim() error {
	if !c.Status.IsEnabled() {
		return valueobject.ErrInvalidCouponStatus
	}

	if err := c.ValidPeriod.Validate(); err != nil {
		return err
	}

	if c.RemainingCount <= 0 {
		return ErrCouponOutOfStock
	}

	return nil
}

// Claim 领取优惠券
func (c *Coupon) Claim() error {
	if err := c.CanClaim(); err != nil {
		return err
	}

	c.RemainingCount--
	c.UpdatedAt = time.Now()
	return nil
}

// ReturnStock 归还库存
func (c *Coupon) ReturnStock() {
	if c.RemainingCount < c.TotalCount {
		c.RemainingCount++
		c.UpdatedAt = time.Now()
	}
}

// CanUse 是否可使用
func (c *Coupon) CanUse(orderAmount int64) error {
	if !c.Status.IsEnabled() {
		return valueobject.ErrInvalidCouponStatus
	}

	if err := c.ValidPeriod.Validate(); err != nil {
		return err
	}

	// 检查金额是否满足最低消费
	_, err := c.Discount.Calculate(orderAmount)
	return err
}

// CalculateDiscount 计算优惠金额
func (c *Coupon) CalculateDiscount(orderAmount int64) (int64, error) {
	return c.Discount.Calculate(orderAmount)
}

// Enable 启用优惠券
func (c *Coupon) Enable() {
	c.Status = valueobject.CouponStatusEnabled
	c.UpdatedAt = time.Now()
}

// Disable 禁用优惠券
func (c *Coupon) Disable() {
	c.Status = valueobject.CouponStatusDisabled
	c.UpdatedAt = time.Now()
}

// IsExpired 是否已过期
func (c *Coupon) IsExpired() bool {
	return c.ValidPeriod.IsExpired()
}

// IsEnabled 是否已开始
func (c *Coupon) IsActive() bool {
	return c.Status.IsEnabled() && c.ValidPeriod.IsActive()
}
