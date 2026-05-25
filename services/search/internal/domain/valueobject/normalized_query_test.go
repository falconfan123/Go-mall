package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizedQuery(t *testing.T) {
	t.Parallel()

	query := NewNormalizedQuery("  iPhone 15 Pro  ")
	require.Equal(t, "iphone 15 pro", query.GetValue())
	require.Equal(t, []string{"iphone", "15", "pro"}, query.Tokens())

	var nilQuery *NormalizedQuery
	require.Empty(t, nilQuery.GetValue())
	require.Empty(t, nilQuery.Tokens())
}
