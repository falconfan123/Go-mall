package aggregate

import (
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestCheckoutAggregateAddDiscountAndCreateFromCart(t *testing.T) {
	t.Parallel()

	checkout := NewCheckoutAggregate("pre-1", 1, 2, 30)
	require.NoError(t, checkout.AddItem(1001, "product", "img", 2, 100))
	require.NoError(t, checkout.AddItem(1002, "product-2", "img2", 1, 200))
	require.Equal(t, int64(400), checkout.CalculateTotal())

	require.NoError(t, checkout.ApplyDiscount(50))
	require.Equal(t, int64(350), checkout.GetFinalAmount())

	created := NewCheckoutAggregate("pre-2", 0, 0, 30)
	err := created.CreateFromCart([]*CartItemInfo{
		{ProductID: 2001, ProductName: "name", ProductImage: "img", Quantity: 2, Price: 300},
	}, &valueobject.Address{ID: 10, UserID: 9}, 30)
	require.NoError(t, err)
	require.Equal(t, int64(600), created.GetOriginalAmount())
	require.Equal(t, int64(9), created.GetUserID())
	require.Equal(t, int64(10), created.GetAddressID())
}

func TestCheckoutAggregateInvalidItemAndExpiration(t *testing.T) {
	t.Parallel()

	checkout := NewCheckoutAggregate("pre-1", 1, 2, 30)
	require.ErrorIs(t, checkout.AddItem(1001, "product", "img", 0, 100), ErrInvalidCheckoutItem)

	entityCheckout := checkout.GetCheckout()
	entityCheckout.ExpireTime = time.Now().Add(-time.Minute)
	require.True(t, checkout.IsExpired())
	require.Error(t, checkout.Confirm())

	err := checkout.CreateFromCart(nil, &valueobject.Address{}, 30)
	require.Error(t, err)
}

func TestCheckoutAggregateStateGetters(t *testing.T) {
	t.Parallel()

	checkout := NewCheckoutAggregate("pre-3", 11, 22, 30)
	require.True(t, checkout.IsPending())
	require.Equal(t, "pre-3", checkout.GetPreOrderID())
	require.Equal(t, int64(11), checkout.GetUserID())
	require.Equal(t, int64(22), checkout.GetAddressID())
	require.Equal(t, int64(0), checkout.GetOriginalAmount())
	require.Len(t, checkout.GetItems(), 0)

	require.NoError(t, checkout.AddItem(1001, "product", "img", 2, 120))
	require.Equal(t, 2, checkout.GetTotalQuantity())
	require.Len(t, checkout.GetItems(), 1)
	require.Greater(t, checkout.GetExpireTime().Unix(), int64(0))

	loaded := LoadCheckout(checkout.GetCheckout())
	require.Same(t, checkout.GetCheckout(), loaded.GetCheckout())
}

func TestCheckoutAggregateCancelAndExpire(t *testing.T) {
	t.Parallel()

	checkout := NewCheckoutAggregate("pre-4", 1, 2, 30)
	require.NoError(t, checkout.Cancel())
	require.False(t, checkout.IsPending())

	expiring := NewCheckoutAggregate("pre-5", 1, 2, 30)
	expiring.Expire()
	require.False(t, expiring.IsPending())
	require.Equal(t, entity.CheckoutStatusExpired, expiring.GetCheckout().Status)
}
