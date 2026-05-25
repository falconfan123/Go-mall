package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscountCalculate(t *testing.T) {
	t.Parallel()

	fullReduction, err := NewDiscount(CouponTypeFullReduction, 500, 1000)
	require.NoError(t, err)
	discountAmount, err := fullReduction.Calculate(1500)
	require.NoError(t, err)
	require.Equal(t, int64(500), discountAmount)

	discountCoupon, err := NewDiscount(CouponTypeDiscount, 90, 0)
	require.NoError(t, err)
	discountAmount, err = discountCoupon.Calculate(2000)
	require.NoError(t, err)
	require.Equal(t, int64(200), discountAmount)

	directReduction, err := NewDiscount(CouponTypeDirectReduction, 5000, 0)
	require.NoError(t, err)
	finalAmount, err := directReduction.CalculateFinalAmount(3000)
	require.NoError(t, err)
	require.Zero(t, finalAmount)
}

func TestDiscountValidateInputAndThreshold(t *testing.T) {
	t.Parallel()

	_, err := NewDiscount(CouponTypeDiscount, 101, 0)
	require.ErrorIs(t, err, ErrInvalidDiscountValue)

	_, err = NewDiscount(CouponTypeFullReduction, 1, -1)
	require.ErrorIs(t, err, ErrMinAmountNegative)

	discount, err := NewDiscount(CouponTypeFullReduction, 100, 1000)
	require.NoError(t, err)
	_, err = discount.Calculate(999)
	require.ErrorIs(t, err, ErrAmountTooLow)
}
