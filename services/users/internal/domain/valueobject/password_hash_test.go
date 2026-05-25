package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPasswordHashVerifyAndEquals(t *testing.T) {
	t.Parallel()

	hash := NewPasswordHash("secret-123")
	require.NotEmpty(t, hash.Value())
	require.True(t, hash.Verify("secret-123"))
	require.False(t, hash.Verify("wrong"))

	same := NewPasswordHashFromHash(hash.Value())
	require.True(t, hash.Equals(*same))
}
