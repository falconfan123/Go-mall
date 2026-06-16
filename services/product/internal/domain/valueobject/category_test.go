package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCategoryValueObject(t *testing.T) {
	t.Parallel()

	category, err := NewCategory(1, "phone")
	require.NoError(t, err)
	require.True(t, category.Equals(Category{ID: 1, Name: "phone"}))
	require.False(t, category.Equals(Category{ID: 2, Name: "phone"}))
	require.Equal(t, category, category.Value())
}

func TestCategoryValidation(t *testing.T) {
	t.Parallel()

	_, err := NewCategory(-1, "phone")
	require.ErrorIs(t, err, ErrInvalidCategoryID)

	_, err = NewCategory(1, "")
	require.ErrorIs(t, err, ErrInvalidCategoryName)
}
