package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntityTagsDeduplicateAndEmpty(t *testing.T) {
	t.Parallel()

	tags := NewEntityTags()
	require.True(t, tags.IsEmpty())

	tags.AddBrand("apple")
	tags.AddBrand("apple")
	tags.AddProductWord("phone")
	tags.AddModifier("blue")

	require.Equal(t, []string{"apple"}, tags.Brands)
	require.Equal(t, []string{"phone"}, tags.ProductWords)
	require.Equal(t, []string{"blue"}, tags.Modifiers)
	require.False(t, tags.IsEmpty())
}
