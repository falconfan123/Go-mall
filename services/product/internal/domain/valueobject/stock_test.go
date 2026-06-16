package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStockValueObject(t *testing.T) {
	t.Parallel()

	stock, err := NewStock(5)
	require.NoError(t, err)
	require.Equal(t, int64(5), stock.Value())
	require.True(t, stock.IsAvailable(3))
	require.False(t, stock.IsAvailable(6))

	adjusted, err := stock.Adjust(-2)
	require.NoError(t, err)
	require.Equal(t, int64(3), adjusted.Value())
}

func TestStockValidation(t *testing.T) {
	t.Parallel()

	_, err := NewStock(-1)
	require.ErrorIs(t, err, ErrInvalidStock)

	stock, err := NewStock(1)
	require.NoError(t, err)

	_, err = stock.Adjust(-2)
	require.ErrorIs(t, err, ErrInsufficientStock)
}
