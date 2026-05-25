package aggregate

import (
	"testing"

	"github.com/falconfan123/Go-mall/services/carts/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestCartAddIncreaseDecreaseAndTotals(t *testing.T) {
	t.Parallel()

	cart := NewCart(1)
	item := newCartItem(t, 1001, 2, 9.9)

	require.NoError(t, cart.AddItem(item))
	require.NoError(t, cart.AddItem(newCartItem(t, 1001, 1, 9.9)))
	require.Equal(t, int32(3), cart.GetTotalQuantity())
	require.InDelta(t, 29.7, cart.GetTotalAmount(), 0.001)

	require.NoError(t, cart.ToggleItemCheck(1001))
	require.Zero(t, cart.GetTotalAmount())

	require.NoError(t, cart.ToggleItemCheck(1001))
	require.NoError(t, cart.DecreaseItemQuantity(1001, 3))
	require.Len(t, cart.Items, 0)
}

func TestCartErrorsAndCheckOperations(t *testing.T) {
	t.Parallel()

	cart := NewCart(1)
	require.ErrorIs(t, cart.RemoveItem(10), ErrItemNotFound)
	require.ErrorIs(t, cart.IncreaseItemQuantity(10, 1), ErrItemNotFound)

	require.NoError(t, cart.AddItem(newCartItem(t, 1001, 1, 10)))
	require.NoError(t, cart.AddItem(newCartItem(t, 1002, 2, 20)))

	cart.UncheckAll()
	require.Empty(t, cart.GetCheckedItems())

	cart.CheckAll()
	require.Len(t, cart.GetCheckedItems(), 2)

	qty, err := cart.GetItemQuantity(1002)
	require.NoError(t, err)
	require.Equal(t, int32(2), qty)
}

func newCartItem(t *testing.T, productID int64, quantity int32, price float64) *entity.CartItem {
	t.Helper()

	qty, err := valueobject.NewQuantity(quantity)
	require.NoError(t, err)
	return entity.NewCartItem(productID, "name", "image", price, qty)
}
