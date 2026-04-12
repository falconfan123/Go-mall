package repository

import (
	"context"
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/aggregate"
)

// InventoryRepository 库存仓储接口
type InventoryRepository interface {
	// Save 保存库存（新建或更新）
	Save(ctx context.Context, inventory *aggregate.Inventory) error

	// GetByProductID 根据商品ID查询库存
	GetByProductID(ctx context.Context, productID int64) (*aggregate.Inventory, error)

	// BatchGetByProductIDs 批量查询库存
	BatchGetByProductIDs(ctx context.Context, productIDs []int64) ([]*aggregate.Inventory, error)

	// SavePreInventoryRecord 保存预扣库存记录
	SavePreInventoryRecord(ctx context.Context, record *aggregate.Inventory) error

	// RemovePreInventoryRecord 删除预扣库存记录
	RemovePreInventoryRecord(ctx context.Context, preOrderID string, productID int64) error

	// GetPreInventoryRecord 根据预订单ID和商品ID查询预扣记录
	GetPreInventoryRecord(ctx context.Context, preOrderID string, productID int64) (*aggregate.Inventory, error)

	// DecreaseStock 原子扣减库存
	DecreaseStock(ctx context.Context, productID int64, quantity int64) error

	// IncreaseStock 原子增加库存
	IncreaseStock(ctx context.Context, productID int64, quantity int64) error
}
