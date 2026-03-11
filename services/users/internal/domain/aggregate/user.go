package aggregate

import (
	"errors"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/valueobject"
	"time"
)

var (
	ErrAddressNotFound = errors.New("address not found")
	ErrTooManyAddresses = errors.New("address count exceeds limit (max 10)")
)

const (
	MaxAddressCount = 10 // 每个用户最多10个收货地址
)

// User 用户聚合根
type User struct {
	ID           int64                     // 用户ID
	Email        *valueobject.Email        // 邮箱
	PasswordHash *valueobject.PasswordHash // 密码哈希
	Username     string                    // 用户名
	Nickname     string                    // 昵称
	Avatar       string                    // 头像
	Phone        string                    // 手机号
	Status       int                       // 用户状态 0:正常 1:禁用
	Addresses    []*entity.Address         // 收货地址列表
	CreateTime   time.Time                 // 创建时间
	UpdateTime   time.Time                 // 更新时间
	LastLoginTime time.Time                // 最后登录时间
	LastLoginIP  string                    // 最后登录IP
}

// NewUser 创建新用户
func NewUser(email *valueobject.Email, passwordHash *valueobject.PasswordHash, username string) *User {
	now := time.Now()
	return &User{
		Email:        email,
		PasswordHash: passwordHash,
		Username:     username,
		Status:       0,
		Addresses:    make([]*entity.Address, 0),
		CreateTime:   now,
		UpdateTime:   now,
	}
}

// VerifyPassword 验证密码
func (u *User) VerifyPassword(plainPassword string) bool {
	return u.PasswordHash.Verify(plainPassword)
}

// UpdatePassword 更新密码
func (u *User) UpdatePassword(newPassword *valueobject.PasswordHash) {
	u.PasswordHash = newPassword
	u.UpdateTime = time.Now()
}

// UpdateProfile 更新用户资料
func (u *User) UpdateProfile(nickname, avatar, phone string) {
	u.Nickname = nickname
	u.Avatar = avatar
	u.Phone = phone
	u.UpdateTime = time.Now()
}

// RecordLogin 记录登录信息
func (u *User) RecordLogin(ip string) {
	u.LastLoginTime = time.Now()
	u.LastLoginIP = ip
	u.UpdateTime = time.Now()
}

// AddAddress 添加收货地址
func (u *User) AddAddress(address *entity.Address) error {
	if len(u.Addresses) >= MaxAddressCount {
		return ErrTooManyAddresses
	}

	// 如果是第一个地址，默认设置为默认地址
	if len(u.Addresses) == 0 {
		address.SetDefault()
	} else if address.IsDefault {
		// 如果新地址是默认地址，取消其他地址的默认状态
		u.cancelAllDefaultAddresses()
	}

	u.Addresses = append(u.Addresses, address)
	u.UpdateTime = time.Now()
	return nil
}

// UpdateAddress 更新收货地址
func (u *User) UpdateAddress(addressID int64, receiver, phone string, addressInfo *valueobject.AddressInfo, isDefault bool) error {
	address, err := u.findAddressByID(addressID)
	if err != nil {
		return err
	}

	if isDefault && !address.IsDefault {
		u.cancelAllDefaultAddresses()
	}

	address.Update(receiver, phone, addressInfo, isDefault)
	u.UpdateTime = time.Now()
	return nil
}

// DeleteAddress 删除收货地址
func (u *User) DeleteAddress(addressID int64) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			u.Addresses = append(u.Addresses[:i], u.Addresses[i+1:]...)
			u.UpdateTime = time.Now()

			// 如果删除的是默认地址，将第一个地址设置为默认
			if addr.IsDefault && len(u.Addresses) > 0 {
				u.Addresses[0].SetDefault()
			}
			return nil
		}
	}
	return ErrAddressNotFound
}

// SetDefaultAddress 设置默认地址
func (u *User) SetDefaultAddress(addressID int64) error {
	address, err := u.findAddressByID(addressID)
	if err != nil {
		return err
	}

	u.cancelAllDefaultAddresses()
	address.SetDefault()
	u.UpdateTime = time.Now()
	return nil
}

// GetDefaultAddress 获取默认地址
func (u *User) GetDefaultAddress() (*entity.Address, error) {
	for _, addr := range u.Addresses {
		if addr.IsDefault {
			return addr, nil
		}
	}
	return nil, ErrAddressNotFound
}

// 辅助方法：取消所有地址的默认状态
func (u *User) cancelAllDefaultAddresses() {
	for _, addr := range u.Addresses {
		if addr.IsDefault {
			addr.CancelDefault()
		}
	}
}

// 辅助方法：根据ID查找地址
func (u *User) findAddressByID(addressID int64) (*entity.Address, error) {
	for _, addr := range u.Addresses {
		if addr.ID == addressID {
			return addr, nil
		}
	}
	return nil, ErrAddressNotFound
}
