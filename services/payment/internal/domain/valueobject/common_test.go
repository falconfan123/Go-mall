package valueobject

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMoneyAndExpireTime(t *testing.T) {
	t.Parallel()

	money := NewMoney(1000)
	require.Equal(t, int64(1200), money.Add(NewMoney(200)).Amount())
	require.Equal(t, int64(700), money.Subtract(NewMoney(300)).Amount())

	expireTime := NewExpireTime(time.Now().Add(time.Minute))
	require.False(t, expireTime.IsExpired())
	require.WithinDuration(t, expireTime.Value(), NewExpireTime(expireTime.Value()).Value(), time.Millisecond)
}

func TestIdentifierValueObjects(t *testing.T) {
	t.Parallel()

	require.Equal(t, "pay-1", NewPaymentID("pay-1").Value())
	require.Equal(t, "pre-1", NewPreOrderID("pre-1").Value())
	require.Equal(t, "order-1", NewOrderID("order-1").Value())

	now := time.Now()
	require.WithinDuration(t, now, NewPaymentTime(now).Value(), time.Millisecond)
}
