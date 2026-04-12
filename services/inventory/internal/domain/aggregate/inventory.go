package aggregate

import (
	"errors"
	"time"

	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/valueobject"
)

// Inventory 库存聚合根
type Inventory struct {
	ProductID   int64                            // 商品ID
	TotalStock  valueobject.Stock                // 总库存
	LockedStock valueobject.Stock                // 已锁定库存（预扣）
	SoldCount   int64                            // 已售数量
	PreRecords  []valueobject.PreInventoryRecord // 预扣记录
	UpdatedAt   time.Time                        // 更新时间
}

// NewInventory 创建新库存
func NewInventory(
	productID int64,
	totalStock valueobject.Stock,
) *Inventory {
	return &Inventory{
		ProductID:   productID,
		TotalStock:  totalStock,
		LockedStock: valueobject.Stock(0),
		SoldCount:   0,
		PreRecords:  make([]valueobject.PreInventoryRecord, 0),
		UpdatedAt:   time.Now(),
	}
}

// AvailableStock 可用库存 = 总库存 - 已锁定库存
func (i *Inventory) AvailableStock() valueobject.Stock {
	available, _ := i.TotalStock.Subtract(i.LockedStock.Value())
	return available
}

// PreDecrease 预扣减库存
func (i *Inventory) PreDecrease(record valueobject.PreInventoryRecord) error {
	// 检查可用库存是否足够
	if !i.AvailableStock().IsAvailable(record.Quantity) {
		return valueobject.ErrInsufficientStock
	}

	// 增加锁定库存
	newLocked, err := i.LockedStock.Add(record.Quantity)
	if err != nil {
		return err
	}
	i.LockedStock = newLocked

	// 添加预扣记录
	i.PreRecords = append(i.PreRecords, record)
	i.UpdatedAt = time.Now()
	return nil
}

// ConfirmDecrease 确认扣减库存（支付成功后调用）
func (i *Inventory) ConfirmDecrease(preOrderID string, productID int64) error {
	// 查找预扣记录
	var foundRecord *valueobject.PreInventoryRecord
	var recordIndex int = -1
	for idx, r := range i.PreRecords {
		if r.PreOrderID == preOrderID && r.ProductID == productID {
			foundRecord = &r
			recordIndex = idx
			break
		}
	}

	if foundRecord == nil {
		return errors.New("pre inventory record not found")
	}

	if foundRecord.IsExpired() {
		return errors.New("pre inventory record has expired")
	}

	// 扣减总库存
	newTotal, err := i.TotalStock.Subtract(foundRecord.Quantity)
	if err != nil {
		return err
	}
	i.TotalStock = newTotal

	// 减少锁定库存
	newLocked, err := i.LockedStock.Subtract(foundRecord.Quantity)
	if err != nil {
		return err
	}
	i.LockedStock = newLocked

	// 增加销量
	i.SoldCount += foundRecord.Quantity

	// 移除预扣记录
	i.PreRecords = append(i.PreRecords[:recordIndex], i.PreRecords[recordIndex+1:]...)
	i.UpdatedAt = time.Now()
	return nil
}

// ReturnPreInventory 退还预扣库存
func (i *Inventory) ReturnPreInventory(preOrderID string, productID int64) error {
	// 查找预扣记录
	var foundRecord *valueobject.PreInventoryRecord
	var recordIndex int = -1
	for idx, r := range i.PreRecords {
		if r.PreOrderID == preOrderID && r.ProductID == productID {
			foundRecord = &r
			recordIndex = idx
			break
		}
	}

	if foundRecord == nil {
		return errors.New("pre inventory record not found")
	}

	// 减少锁定库存
	newLocked, err := i.LockedStock.Subtract(foundRecord.Quantity)
	if err != nil {
		return err
	}
	i.LockedStock = newLocked

	// 移除预扣记录
	i.PreRecords = append(i.PreRecords[:recordIndex], i.PreRecords[recordIndex+1:]...)
	i.UpdatedAt = time.Now()
	return nil
}

// DirectDecrease 直接扣减库存（无需预扣）
func (i *Inventory) DirectDecrease(quantity int64) error {
	if !i.AvailableStock().IsAvailable(quantity) {
		return valueobject.ErrInsufficientStock
	}

	// 扣减总库存
	newTotal, err := i.TotalStock.Subtract(quantity)
	if err != nil {
		return err
	}
	i.TotalStock = newTotal

	// 增加销量
	i.SoldCount += quantity
	i.UpdatedAt = time.Now()
	return nil
}

// ReturnInventory 直接退还库存（取消订单时调用）
func (i *Inventory) ReturnInventory(quantity int64) error {
	// 增加总库存
	newTotal, err := i.TotalStock.Add(quantity)
	if err != nil {
		return err
	}
	i.TotalStock = newTotal

	// 减少销量
	if i.SoldCount >= quantity {
		i.SoldCount -= quantity
	}
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateStock 直接更新库存数量
func (i *Inventory) UpdateStock(newStock valueobject.Stock) {
	i.TotalStock = newStock
	i.UpdatedAt = time.Now()
}

// CleanExpiredPreRecords 清理过期的预扣记录
func (i *Inventory) CleanExpiredPreRecords() int64 {
	validRecords := make([]valueobject.PreInventoryRecord, 0)
	var releasedQuantity int64 = 0

	for _, r := range i.PreRecords {
		if r.IsExpired() {
			releasedQuantity += r.Quantity
		} else {
			validRecords = append(validRecords, r)
		}
	}

	if releasedQuantity > 0 {
		// 释放锁定库存
		newLocked, _ := i.LockedStock.Subtract(releasedQuantity)
		i.LockedStock = newLocked
		i.PreRecords = validRecords
		i.UpdatedAt = time.Now()
	}

	return releasedQuantity
}
