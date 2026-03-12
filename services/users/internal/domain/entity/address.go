package entity

import (
	"github.com/falconfan123/Go-mall/services/users/internal/domain/valueobject"
	"time"
)

// Address 收货地址实体
type Address struct {
	ID         int64                    // 地址ID
	UserID     int64                    // 所属用户ID
	Receiver   string                   // 收货人
	Phone      string                   // 联系电话
	Address    *valueobject.AddressInfo // 地址详情
	IsDefault  bool                     // 是否默认地址
	CreateTime time.Time                // 创建时间
	UpdateTime time.Time                // 更新时间
}

// NewAddress 创建新地址
func NewAddress(userID int64, receiver, phone string, address *valueobject.AddressInfo, isDefault bool) *Address {
	now := time.Now()
	return &Address{
		UserID:     userID,
		Receiver:   receiver,
		Phone:      phone,
		Address:    address,
		IsDefault:  isDefault,
		CreateTime: now,
		UpdateTime: now,
	}
}

// Update 更新地址信息
func (a *Address) Update(receiver, phone string, address *valueobject.AddressInfo, isDefault bool) {
	a.Receiver = receiver
	a.Phone = phone
	a.Address = address
	a.IsDefault = isDefault
	a.UpdateTime = time.Now()
}

// SetDefault 设置为默认地址
func (a *Address) SetDefault() {
	a.IsDefault = true
	a.UpdateTime = time.Now()
}

// CancelDefault 取消默认地址
func (a *Address) CancelDefault() {
	a.IsDefault = false
	a.UpdateTime = time.Now()
}

// Equals 比较两个地址是否相同
func (a *Address) Equals(other *Address) bool {
	if a == nil || other == nil {
		return false
	}
	return a.Receiver == other.Receiver &&
		a.Phone == other.Phone &&
		a.Address.Equals(*other.Address)
}
