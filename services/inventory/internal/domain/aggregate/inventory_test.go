package aggregate

import (
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestInventoryPreDecreaseConfirmAndReturn(t *testing.T) {
	t.Parallel()

	total, err := valueobject.NewStock(10)
	require.NoError(t, err)
	inventory := NewInventory(1001, total)

	record, err := valueobject.NewPreInventoryRecord(1001, 3, "pre-1", 10, time.Now().Add(time.Hour))
	require.NoError(t, err)
	require.NoError(t, inventory.PreDecrease(record))
	require.Equal(t, int64(7), inventory.AvailableStock().Value())

	require.NoError(t, inventory.ConfirmDecrease("pre-1", 1001))
	require.Equal(t, int64(7), inventory.TotalStock.Value())
	require.Equal(t, int64(0), inventory.LockedStock.Value())
	require.Equal(t, int64(3), inventory.SoldCount)

	record2, err := valueobject.NewPreInventoryRecord(1001, 2, "pre-2", 10, time.Now().Add(time.Hour))
	require.NoError(t, err)
	require.NoError(t, inventory.PreDecrease(record2))
	require.NoError(t, inventory.ReturnPreInventory("pre-2", 1001))
	require.Equal(t, int64(0), inventory.LockedStock.Value())
}

func TestInventoryDirectDecreaseAndCleanExpired(t *testing.T) {
	t.Parallel()

	total, err := valueobject.NewStock(5)
	require.NoError(t, err)
	inventory := NewInventory(1001, total)

	require.NoError(t, inventory.DirectDecrease(2))
	require.Equal(t, int64(3), inventory.TotalStock.Value())
	require.Equal(t, int64(2), inventory.SoldCount)

	expiredRecord := valueobject.PreInventoryRecord{
		ProductID:  1001,
		Quantity:   1,
		PreOrderID: "expired",
		UserID:     10,
		ExpireTime: time.Now().Add(-time.Minute),
	}
	require.NoError(t, inventory.PreDecrease(valueobject.PreInventoryRecord{
		ProductID:  1001,
		Quantity:   1,
		PreOrderID: "locked",
		UserID:     10,
		ExpireTime: time.Now().Add(time.Hour),
	}))
	inventory.PreRecords = append(inventory.PreRecords, expiredRecord)
	inventory.LockedStock = valueobject.Stock(2)

	released := inventory.CleanExpiredPreRecords()
	require.Equal(t, int64(1), released)
	require.Equal(t, int64(1), inventory.LockedStock.Value())

	require.NoError(t, inventory.ReturnInventory(1))
	require.Equal(t, int64(4), inventory.TotalStock.Value())
}
