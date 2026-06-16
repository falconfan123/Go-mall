package valueobject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddressInfoValueObject(t *testing.T) {
	t.Parallel()

	address := NewAddressInfo("Zhejiang", "Hangzhou", "Xihu", "No.1 Road", "310000")
	require.False(t, address.IsEmpty())
	require.Equal(t, "ZhejiangHangzhouXihuNo.1 Road", address.FullAddress())
	require.True(t, address.Equals(AddressInfo{
		Province: "Zhejiang",
		City:     "Hangzhou",
		District: "Xihu",
		Detail:   "No.1 Road",
		ZipCode:  "310000",
	}))
}

func TestAddressInfoIsEmpty(t *testing.T) {
	t.Parallel()

	address := NewAddressInfo("", "Hangzhou", "Xihu", "No.1 Road", "310000")
	require.True(t, address.IsEmpty())
}
