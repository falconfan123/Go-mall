package aggregate

import (
	"testing"

	"github.com/falconfan123/Go-mall/services/users/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestUserAddressManagement(t *testing.T) {
	t.Parallel()

	email, err := valueobject.NewEmail("user@example.com")
	require.NoError(t, err)
	user := NewUser(email, valueobject.NewPasswordHash("secret-123"), "tester")

	first := newAddress(t, 1, false)
	require.NoError(t, user.AddAddress(first))
	require.True(t, first.IsDefault)

	second := newAddress(t, 2, true)
	require.NoError(t, user.AddAddress(second))
	require.False(t, first.IsDefault)
	require.True(t, second.IsDefault)

	err = user.SetDefaultAddress(1)
	require.NoError(t, err)
	require.True(t, first.IsDefault)
	require.False(t, second.IsDefault)

	defaultAddress, err := user.GetDefaultAddress()
	require.NoError(t, err)
	require.Equal(t, int64(1), defaultAddress.ID)

	err = user.DeleteAddress(1)
	require.NoError(t, err)
	require.Len(t, user.Addresses, 1)
	require.True(t, user.Addresses[0].IsDefault)
}

func TestUserAddressLimitAndUpdateMissing(t *testing.T) {
	t.Parallel()

	user := NewUser(nil, valueobject.NewPasswordHash("secret-123"), "tester")
	for i := 0; i < MaxAddressCount; i++ {
		require.NoError(t, user.AddAddress(newAddress(t, int64(i+1), false)))
	}

	err := user.AddAddress(newAddress(t, 99, false))
	require.ErrorIs(t, err, ErrTooManyAddresses)

	info := valueobject.NewAddressInfo("p", "c", "d", "detail", "zip")
	err = user.UpdateAddress(1000, "receiver", "123", info, false)
	require.ErrorIs(t, err, ErrAddressNotFound)
}

func TestUserVerifyAndUpdatePassword(t *testing.T) {
	t.Parallel()

	user := NewUser(nil, valueobject.NewPasswordHash("old-pass"), "tester")
	require.True(t, user.VerifyPassword("old-pass"))

	newHash := valueobject.NewPasswordHash("new-pass")
	user.UpdatePassword(newHash)

	require.True(t, user.VerifyPassword("new-pass"))
	require.False(t, user.VerifyPassword("old-pass"))
}

func newAddress(t *testing.T, id int64, isDefault bool) *entity.Address {
	t.Helper()

	info := valueobject.NewAddressInfo("Zhejiang", "Hangzhou", "Xihu", "No.1", "310000")
	address := entity.NewAddress(1, "fan", "13800138000", info, isDefault)
	address.ID = id
	return address
}
