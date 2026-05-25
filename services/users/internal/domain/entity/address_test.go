package entity

import (
	"testing"

	"github.com/falconfan123/Go-mall/services/users/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestAddressEntityBehaviors(t *testing.T) {
	t.Parallel()

	info := valueobject.NewAddressInfo("Zhejiang", "Hangzhou", "Xihu", "No.1", "310000")
	address := NewAddress(1, "fan", "13800138000", info, false)
	require.False(t, address.IsDefault)

	address.SetDefault()
	require.True(t, address.IsDefault)

	address.CancelDefault()
	require.False(t, address.IsDefault)

	address.Update("new", "13900139000", info, true)
	require.Equal(t, "new", address.Receiver)
	require.True(t, address.IsDefault)

	same := NewAddress(1, "new", "13900139000", info, true)
	require.True(t, address.Equals(same))
}
