package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenInfoAndAuthResult(t *testing.T) {
	t.Parallel()

	token := NewTokenInfo(1, "127.0.0.1", time.Now().Add(-time.Minute), time.Now().Add(time.Minute))
	require.False(t, token.IsExpired())

	expired := NewTokenInfo(1, "127.0.0.1", time.Now().Add(-2*time.Minute), time.Now().Add(-time.Minute))
	require.True(t, expired.IsExpired())

	result := NewAuthResult(1, 0, "ok", true)
	require.True(t, result.IsValid)
	require.Equal(t, "ok", result.StatusMsg)
}
