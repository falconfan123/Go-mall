package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMoneyAndPreOrderID(t *testing.T) {
	t.Parallel()

	money := NewMoney(1000)
	require.Equal(t, int64(1500), money.Add(NewMoney(500)).Amount)
	require.Equal(t, int64(500), money.Subtract(NewMoney(500)).Amount)
	require.Equal(t, int64(2000), money.Multiply(2).Amount)
	require.InDelta(t, 10.0, money.ToYuan(), 0.001)

	id := NewPreOrderID(1)
	require.NotEmpty(t, id.Value)
}
