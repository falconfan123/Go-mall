package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewQuantity(t *testing.T) {
	t.Parallel()

	qty, err := NewQuantity(1)
	require.NoError(t, err)
	require.Equal(t, int32(1), qty.Value())

	_, err = NewQuantity(0)
	require.ErrorIs(t, err, ErrInvalidQuantity)

	_, err = NewQuantity(MaxQuantity + 1)
	require.ErrorIs(t, err, ErrMaxQuantityExceeded)
}

func TestQuantityAddAndSubtract(t *testing.T) {
	t.Parallel()

	qty, err := NewQuantity(2)
	require.NoError(t, err)

	increased, err := qty.Add(3)
	require.NoError(t, err)
	require.Equal(t, int32(5), increased.Value())

	decreased, err := increased.Subtract(5)
	require.NoError(t, err)
	require.Equal(t, int32(0), decreased.Value())

	_, err = increased.Add(MaxQuantity)
	require.ErrorIs(t, err, ErrMaxQuantityExceeded)

	_, err = qty.Subtract(3)
	require.ErrorIs(t, err, ErrInvalidQuantity)
}
