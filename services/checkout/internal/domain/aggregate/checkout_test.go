package aggregate

import (
	"testing"
	"time"

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
