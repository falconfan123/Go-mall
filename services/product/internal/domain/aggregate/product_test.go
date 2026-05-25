package aggregate

import (
	"testing"

	"github.com/falconfan123/Go-mall/services/product/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestProductBehaviors(t *testing.T) {
	t.Parallel()

	price, err := valueobject.NewPrice(1000)
	require.NoError(t, err)
	stock, err := valueobject.NewStock(5)
	require.NoError(t, err)
	category, err := valueobject.NewCategory(1, "phone")
	require.NoError(t, err)

	product := NewProduct("name", "desc", "pic", price, stock, []valueobject.Category{category}, "thumb")
	require.True(t, product.IsOnSale())
	require.True(t, product.HasCategory(category))

	require.NoError(t, product.AdjustStock(-3))
	require.Equal(t, int64(2), product.Stock.Value())

	product.IncreaseSold(3)
	require.Equal(t, int64(3), product.Sold)

	newPrice, err := valueobject.NewPrice(1500)
	require.NoError(t, err)
	product.UpdateInfo("new", "new-desc", "new-pic", newPrice, []valueobject.Category{category}, "new-thumb")
	require.Equal(t, "new", product.Name)
	require.Equal(t, int64(1500), product.Price.Value())
}

func TestProductAdjustStockRejectsInsufficientInventory(t *testing.T) {
	t.Parallel()

	price, err := valueobject.NewPrice(1000)
	require.NoError(t, err)
	stock, err := valueobject.NewStock(1)
	require.NoError(t, err)

	product := NewProduct("name", "desc", "pic", price, stock, nil, "")
	err = product.AdjustStock(-2)
	require.ErrorIs(t, err, valueobject.ErrInsufficientStock)
}
