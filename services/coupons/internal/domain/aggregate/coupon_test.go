package aggregate

import (
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestCouponClaimUseAndState(t *testing.T) {
	t.Parallel()

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)
	coupon, err := NewCoupon("cp-1", "new", valueobject.CouponTypeFullReduction, 100, 0, start, end, 2)
	require.NoError(t, err)

	require.NoError(t, coupon.CanClaim())
	require.NoError(t, coupon.Claim())
	require.Equal(t, uint64(1), coupon.RemainingCount)

	discount, err := coupon.CalculateDiscount(1000)
	require.NoError(t, err)
	require.Equal(t, int64(100), discount)
	require.NoError(t, coupon.CanUse(1000))

	coupon.Disable()
	require.False(t, coupon.IsActive())
	require.ErrorIs(t, coupon.CanClaim(), valueobject.ErrInvalidCouponStatus)

	coupon.Enable()
	coupon.ReturnStock()
	require.Equal(t, uint64(2), coupon.RemainingCount)
}

func TestCouponRejectsOutOfStockAndNotStarted(t *testing.T) {
	t.Parallel()

	coupon, err := NewCoupon(
		"cp-2",
		"future",
		valueobject.CouponTypeFullReduction,
		100,
		0,
		time.Now().Add(time.Hour),
		time.Now().Add(2*time.Hour),
		0,
	)
	require.NoError(t, err)
	require.ErrorIs(t, coupon.CanClaim(), valueobject.ErrCouponNotActive)

	activeCoupon, err := NewCoupon(
		"cp-3",
		"active",
		valueobject.CouponTypeFullReduction,
		100,
		0,
		time.Now().Add(-time.Hour),
		time.Now().Add(time.Hour),
		0,
	)
	require.NoError(t, err)
	require.ErrorIs(t, activeCoupon.CanClaim(), ErrCouponOutOfStock)
}
