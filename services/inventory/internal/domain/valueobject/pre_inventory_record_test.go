package valueobject

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewPreInventoryRecordAndEquality(t *testing.T) {
	t.Parallel()

	expireAt := time.Now().Add(time.Minute)
	record, err := NewPreInventoryRecord(1, 2, "pre-1", 3, expireAt)
	require.NoError(t, err)
	require.False(t, record.IsExpired())
	require.True(t, record.Equals(record))
}

func TestNewPreInventoryRecordValidation(t *testing.T) {
	t.Parallel()

	_, err := NewPreInventoryRecord(-1, 1, "pre", 1, time.Now().Add(time.Minute))
	require.ErrorIs(t, err, ErrInvalidProductID)

	_, err = NewPreInventoryRecord(1, 0, "pre", 1, time.Now().Add(time.Minute))
	require.ErrorIs(t, err, ErrInvalidQuantity)

	_, err = NewPreInventoryRecord(1, 1, "", 1, time.Now().Add(time.Minute))
	require.ErrorIs(t, err, ErrEmptyPreOrderID)

	_, err = NewPreInventoryRecord(1, 1, "pre", 1, time.Now().Add(-time.Minute))
	require.ErrorIs(t, err, ErrInvalidExpireTime)
}
