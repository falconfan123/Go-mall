package biz

import (
	"errors"
	"time"
)

const (
	InventoryRpcPort = 10011
)
const (
	InventoryKeyPrefix        = "inventory:%d"
	InventoryDeductLockPrefix = "inventory:deduct:lock"
	InventoryProductKey       = "inventory:product"

	// 库存缓存 TTL - 5分钟，数据变更时会主动更新
	InventoryCacheTTL = 5 * time.Minute
)

var (
	// ErrInventoryNotEnough 库存不足err
	ErrInventoryNotEnough = errors.New("not enough inventory")
	// ErrInventoryDecreaseFailed 扣减失败
	ErrInventoryDecreaseFailed = errors.New("decrease inventory failed")
	// ErrInvalidInventory 非法的库存信息
	ErrInvalidInventory = errors.New("invalid inventory")
)
