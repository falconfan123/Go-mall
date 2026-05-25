package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCheckoutEntityLifecycle(t *testing.T) {
	t.Parallel()

	checkout := NewCheckout("pre-1", 1, 10, 0, 30)
	checkout.AddItem(NewCheckoutItem("pre-1", 1001, 2, 300, &ProductSnapshot{ProductName: "product"}))
	checkout.AddItem(NewCheckoutItem("pre-1", 1002, 1, 100, &ProductSnapshot{ProductName: "product-2"}))

	require.Equal(t, int64(700), checkout.CalculateTotal())
	require.NoError(t, checkout.ApplyDiscount(200))
	require.Equal(t, int64(500), checkout.FinalAmount)
	require.NoError(t, checkout.Confirm())
	require.Equal(t, CheckoutStatusConfirmed, checkout.Status)
}

func TestCheckoutEntityExpirationAndCancelRules(t *testing.T) {
	t.Parallel()

	checkout := NewCheckout("pre-2", 1, 10, 0, 30)
	checkout.ExpireTime = time.Now().Add(-time.Minute)
	require.ErrorIs(t, checkout.Confirm(), ErrCheckoutExpired)
	require.Equal(t, CheckoutStatusExpired, checkout.Status)

	confirmed := NewCheckout("pre-3", 1, 10, 0, 30)
	require.NoError(t, confirmed.Confirm())
	require.ErrorIs(t, confirmed.Cancel(), ErrCheckoutAlreadyUsed)
}
