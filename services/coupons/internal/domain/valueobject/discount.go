package valueobject

import (
	"errors"
)

// Discount 折扣值对象
type Discount struct {
	couponType CouponType
	value      int64 // 优惠值：满减/立减单位为分，折扣为百分比*100（如9折为90）
	minAmount  int64 // 最低消费金额（分）
}

var (
	ErrInvalidDiscountValue = errors.New("invalid discount value")
	ErrMinAmountNegative    = errors.New("min amount cannot be negative")
	ErrAmountTooLow         = errors.New("order amount too low to use this coupon")
)

// NewDiscount 创建折扣对象
func NewDiscount(couponType CouponType, value int64, minAmount int64) (*Discount, error) {
	if !couponType.IsValid() {
		return nil, ErrInvalidCouponType
	}

	if value <= 0 {
		return nil, ErrInvalidDiscountValue
	}

	// 折扣类型，值必须在1-100之间
	if couponType == CouponTypeDiscount && (value < 1 || value > 100) {
		return nil, ErrInvalidDiscountValue
	}

	if minAmount < 0 {
		return nil, ErrMinAmountNegative
	}

	return &Discount{
		couponType: couponType,
		value:      value,
		minAmount:  minAmount,
	}, nil
}

// Calculate 计算优惠金额
func (d *Discount) Calculate(orderAmount int64) (int64, error) {
	if orderAmount < d.minAmount {
		return 0, ErrAmountTooLow
	}

	switch d.couponType {
	case CouponTypeFullReduction:
		// 满减
		return d.value, nil
	case CouponTypeDiscount:
		// 折扣：value是百分比*100，比如9折是90，优惠金额 = 订单金额 * (100 - 折扣)/100
		discountAmount := orderAmount * (100 - d.value) / 100
		return discountAmount, nil
	case CouponTypeDirectReduction:
		// 立减，最多减到0
		if d.value > orderAmount {
			return orderAmount, nil
		}
		return d.value, nil
	default:
		return 0, ErrInvalidCouponType
	}
}

// CalculateFinalAmount 计算最终支付金额
func (d *Discount) CalculateFinalAmount(orderAmount int64) (int64, error) {
	discountAmount, err := d.Calculate(orderAmount)
	if err != nil {
		return orderAmount, err
	}

	finalAmount := orderAmount - discountAmount
	if finalAmount < 0 {
		return 0, nil
	}
	return finalAmount, nil
}

// CouponType 获取优惠券类型
func (d *Discount) CouponType() CouponType {
	return d.couponType
}

// Value 获取优惠值
func (d *Discount) Value() int64 {
	return d.value
}

// MinAmount 获取最低消费金额
func (d *Discount) MinAmount() int64 {
	return d.minAmount
}
