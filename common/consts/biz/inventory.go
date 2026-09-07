package biz

import (
	"errors"
)

const (
	InventoryRpcPort = 10007
)
const (
	InventoryKeyPrefix        = "inventory:%d"
	InventoryDeductLockPrefix = "inventory:deduct:lock"
	InventoryProductKey       = "inventory:product"
)

var (
	// ErrInventoryNotEnough 库存不足err
	ErrInventoryNotEnough = errors.New("not enough inventory")
	// ErrInventoryDecreaseFailed 扣减失败
	ErrInventoryDecreaseFailed = errors.New("decrease inventory failed")
	// ErrReturnAlreadyLocked 回补已被锁定（幂等容忍信号：该订单的回补已在进行/已完成）
	ErrReturnAlreadyLocked = errors.New("return already locked")
	// ErrInvalidInventory 非法的库存信息
	ErrInvalidInventory = errors.New("invalid inventory")
)
