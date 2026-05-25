package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDSLFragmentToMapQuery(t *testing.T) {
	t.Parallel()

	tags := NewEntityTags()
	tags.AddBrand("apple")
	tags.AddProductWord("phone")

	fragment := NewDSLFragment("iphone", tags)
	require.Len(t, fragment.Fields, 4)

	query := fragment.ToMapQuery()
	multiMatch, ok := query["multi_match"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "iphone", multiMatch["query"])
	require.Equal(t, float64(1), multiMatch["boost"])

	fields, ok := multiMatch["fields"].([]string)
	require.True(t, ok)
	require.Contains(t, fields, "name^1")
	require.Contains(t, fields, "brand^2")
	require.Contains(t, fields, "description^0.5")
	require.Contains(t, fields, "category^1.5")
}

func TestDSLFragmentNil(t *testing.T) {
	t.Parallel()

	var fragment *DSLFragment
	require.Nil(t, fragment.ToMapQuery())
}
