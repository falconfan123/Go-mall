package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/falconfan123/Go-mall/dal/model/inventory"
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/valueobject"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// InventoryRepositoryImpl 库存仓储实现
type InventoryRepositoryImpl struct {
	inventoryModel inventory.InventoryModel
	conn           sqlx.SqlConn
}

// NewInventoryRepositoryImpl 创建库存仓储实现
func NewInventoryRepositoryImpl(conn sqlx.SqlConn) repository.InventoryRepository {
	return &InventoryRepositoryImpl{
		inventoryModel: inventory.NewInventoryModel(conn),
		conn:           conn,
	}
}

// Save 保存库存
func (r *InventoryRepositoryImpl) Save(ctx context.Context, inv *aggregate.Inventory) error {
	// 转换领域模型到数据模型
	invData := &inventory.Inventory{
		ProductId: inv.ProductID,
		Total:     inv.TotalStock.Value(),
		Sold:      inv.SoldCount,
	}

	// 检查是否存在
	existing, err := r.inventoryModel.FindOne(ctx, inv.ProductID)
	if err != nil && err != inventory.ErrNotFound {
		return err
	}

	if existing == nil {
		// 新建
		_, err := r.inventoryModel.Insert(ctx, invData)
		return err
	}

	// 更新
	return r.inventoryModel.Update(ctx, invData)
}

// GetByProductID 根据商品ID查询库存
func (r *InventoryRepositoryImpl) GetByProductID(ctx context.Context, productID int64) (*aggregate.Inventory, error) {
	invData, err := r.inventoryModel.FindOne(ctx, productID)
	if err != nil {
		if err == inventory.ErrNotFound {
			// 如果不存在，返回0库存
			stock, _ := valueobject.NewStock(0)
			return aggregate.NewInventory(productID, stock), nil
		}
		return nil, err
	}

	// 转换数据模型到领域模型
	totalStock, err := valueobject.NewStock(invData.Total)
	if err != nil {
		return nil, err
	}

	// 查询预扣库存记录（简化处理，实际应该从预扣表查询）
	lockedStock, _ := valueobject.NewStock(0)
	preRecords := make([]valueobject.PreInventoryRecord, 0)

	return &aggregate.Inventory{
		ProductID:   invData.ProductId,
		TotalStock:  totalStock,
		LockedStock: lockedStock,
		SoldCount:   invData.Sold,
		PreRecords:  preRecords,
		UpdatedAt:   time.Now(),
	}, nil
}

// BatchGetByProductIDs 批量查询库存
func (r *InventoryRepositoryImpl) BatchGetByProductIDs(ctx context.Context, productIDs []int64) ([]*aggregate.Inventory, error) {
	if len(productIDs) == 0 {
		return []*aggregate.Inventory{}, nil
	}

	// 构建IN查询
	placeholders := make([]string, len(productIDs))
	args := make([]interface{}, len(productIDs))
	for i, id := range productIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT product_id, total, sold FROM inventory WHERE product_id IN (%s)", strings.Join(placeholders, ","))
	var invs []*inventory.Inventory
	err := r.conn.QueryRowsCtx(ctx, &invs, query, args...)
	if err != nil {
		return nil, err
	}

	// 转换为领域模型
	result := make([]*aggregate.Inventory, 0, len(invs))
	for _, inv := range invs {
		totalStock, err := valueobject.NewStock(inv.Total)
		if err != nil {
			continue
		}
		lockedStock, _ := valueobject.NewStock(0)
		result = append(result, &aggregate.Inventory{
			ProductID:   inv.ProductId,
			TotalStock:  totalStock,
			LockedStock: lockedStock,
			SoldCount:   inv.Sold,
			PreRecords:  []valueobject.PreInventoryRecord{},
			UpdatedAt:   time.Now(),
		})
	}

	return result, nil
}

// SavePreInventoryRecord 保存预扣库存记录
func (r *InventoryRepositoryImpl) SavePreInventoryRecord(ctx context.Context, record *aggregate.Inventory) error {
	// 简化实现，实际应该保存到预扣库存表
	// 这里省略具体实现，根据业务需求完善
	return nil
}

// RemovePreInventoryRecord 删除预扣库存记录
func (r *InventoryRepositoryImpl) RemovePreInventoryRecord(ctx context.Context, preOrderID string, productID int64) error {
	// 简化实现，实际应该从预扣库存表删除
	return nil
}

// GetPreInventoryRecord 根据预订单ID和商品ID查询预扣记录
func (r *InventoryRepositoryImpl) GetPreInventoryRecord(ctx context.Context, preOrderID string, productID int64) (*aggregate.Inventory, error) {
	// 简化实现，实际应该从预扣库存表查询
	stock, _ := valueobject.NewStock(0)
	return aggregate.NewInventory(productID, stock), nil
}

// DecreaseStock 原子扣减库存
func (r *InventoryRepositoryImpl) DecreaseStock(ctx context.Context, productID int64, quantity int64) error {
	query := "UPDATE inventory SET total = total - ?, sold = sold + ? WHERE product_id = ? AND total >= ?"
	result, err := r.conn.ExecCtx(ctx, query, quantity, quantity, productID, quantity)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return valueobject.ErrInsufficientStock
	}
	return nil
}

// IncreaseStock 原子增加库存
func (r *InventoryRepositoryImpl) IncreaseStock(ctx context.Context, productID int64, quantity int64) error {
	query := "UPDATE inventory SET total = total + ? WHERE product_id = ?"
	_, err := r.conn.ExecCtx(ctx, query, quantity, productID)
	return err
}
