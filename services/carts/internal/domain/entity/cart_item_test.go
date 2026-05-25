package entity

import (
	"testing"

	"github.com/falconfan123/Go-mall/services/carts/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestCartItemQuantityAndCheck(t *testing.T) {
	t.Parallel()

	qty, err := valueobject.NewQuantity(1)
	require.NoError(t, err)
	item := NewCartItem(1, "name", "img", 9.9, qty)

	require.NoError(t, item.IncreaseQuantity(2))
	require.Equal(t, int32(3), item.Quantity.Value())

	require.NoError(t, item.DecreaseQuantity(1))
	require.Equal(t, int32(2), item.Quantity.Value())

	item.ToggleCheck()
	require.False(t, item.Checked)
	item.SetChecked(true)
	require.True(t, item.Checked)
}
