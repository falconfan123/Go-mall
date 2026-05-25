package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPriceOperations(t *testing.T) {
	t.Parallel()

	price, err := NewPrice(1299)
	require.NoError(t, err)
	require.Equal(t, int64(1299), price.Value())
	require.InDelta(t, 12.99, price.ToYuan(), 0.001)

	total := price.Add(Price(200)).Multiply(2)
	require.Equal(t, int64(2998), total.Value())
	require.Equal(t, int64(0), price.Subtract(Price(2000)).Value())
}

func TestNewPriceRejectsNegativeValue(t *testing.T) {
	t.Parallel()

	_, err := NewPrice(-1)
	require.ErrorIs(t, err, ErrInvalidPrice)
}
