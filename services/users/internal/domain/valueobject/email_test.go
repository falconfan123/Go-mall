package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEmail(t *testing.T) {
	t.Parallel()

	email, err := NewEmail("user@example.com")
	require.NoError(t, err)
	require.Equal(t, "user@example.com", email.Value())
	require.Equal(t, "user@example.com", email.String())

	_, err = NewEmail("invalid-email")
	require.ErrorIs(t, err, ErrInvalidEmailFormat)
}

func TestEmailEquals(t *testing.T) {
	t.Parallel()

	left, err := NewEmail("same@example.com")
	require.NoError(t, err)
	right, err := NewEmail("same@example.com")
	require.NoError(t, err)
	other, err := NewEmail("other@example.com")
	require.NoError(t, err)

	require.True(t, left.Equals(*right))
	require.False(t, left.Equals(*other))
}
