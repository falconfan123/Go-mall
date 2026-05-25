package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAuditLog(t *testing.T) {
	t.Parallel()

	log := NewAuditLog(1, "update", "desc", "old", "new", "users", "users", 10, "127.0.0.1", "trace", "span")
	require.EqualValues(t, 1, log.UserID)
	require.Equal(t, "update", log.ActionType)
	require.Equal(t, "users", log.ServiceName)
	require.False(t, log.CreatedAt.IsZero())
}
