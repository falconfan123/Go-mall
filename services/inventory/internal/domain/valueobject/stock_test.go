package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStockOperations(t *testing.T) {
	t.Parallel()

	stock, err := NewStock(10)
	require.NoError(t, err)
	require.True(t, stock.IsAvailable(5))

	added, err := stock.Add(5)
	require.NoError(t, err)
	require.Equal(t, int64(15), added.Value())

	subtracted, err := added.Subtract(7)
	require.NoError(t, err)
	require.Equal(t, int64(8), subtracted.Value())

	_, err = stock.Subtract(11)
	require.ErrorIs(t, err, ErrInsufficientStock)
}

func TestNewStockRejectsNegative(t *testing.T) {
	t.Parallel()

	_, err := NewStock(-1)
	require.ErrorIs(t, err, ErrInvalidStock)
}
