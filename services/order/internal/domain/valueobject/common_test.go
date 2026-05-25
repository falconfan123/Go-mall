package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMoneyOperations(t *testing.T) {
	t.Parallel()

	money := NewMoney(1000)
	require.Equal(t, int64(1200), money.Add(NewMoney(200)).Amount)
	require.Equal(t, int64(700), money.Subtract(NewMoney(300)).Amount)
	require.InDelta(t, 10.0, money.ToYuan(), 0.001)

	orderID := NewOrderID()
	require.NotEmpty(t, orderID.Value)
}
