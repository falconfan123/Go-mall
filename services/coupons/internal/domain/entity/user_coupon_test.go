package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserCouponLifecycle(t *testing.T) {
	t.Parallel()

	coupon := NewUserCoupon(1, "coupon-1")
	require.True(t, coupon.IsUnused())
	require.NoError(t, coupon.CanUse())

	require.NoError(t, coupon.Use("ord-1"))
	require.True(t, coupon.IsUsed())
	require.ErrorIs(t, coupon.CanUse(), ErrUserCouponAlreadyUsed)

	require.NoError(t, coupon.CancelUse())
	require.True(t, coupon.IsUnused())

	coupon.Expire()
	require.True(t, coupon.IsExpired())
	require.ErrorIs(t, coupon.CanUse(), ErrUserCouponExpired)
}
