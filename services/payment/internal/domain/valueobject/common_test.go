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
}
