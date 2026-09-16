package svc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	inventorymodel "github.com/falconfan123/Go-mall/dal/model/inventory"
)

func inventoryCacheKey(productID int64) string {
	return fmt.Sprintf("%s:%d", biz.InventoryProductKey, productID)
}

func (s *ServiceContext) SetInventoryCacheCtx(ctx context.Context, productID, total int64) error {
	return s.Rdb.SetCtx(ctx, inventoryCacheKey(productID), strconv.FormatInt(total, 10))
}

func (s *ServiceContext) GetInventoryCacheCtx(ctx context.Context, productID int64) (int64, bool, error) {
	key := inventoryCacheKey(productID)
	exists, err := s.Rdb.ExistsCtx(ctx, key)
	if err != nil {
		return 0, false, err
	}
	if !exists {
		return 0, false, nil
	}

	raw, err := s.Rdb.GetCtx(ctx, key)
	if err != nil {
		return 0, false, err
	}

	total, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("parse inventory cache %s: %w", key, err)
	}

	return total, true, nil
}

// LoadInventoryFromDBToCache 回填库存缓存。
//
// 缓存键 inventory:product:{pid} 的口径为**可用库存**（可售）。
// DB 仅有 total/sold 两列、无预留列，预扣量只存在于缓存（decreaselua DECRBY），
// 因此回填时无法从 DB 还原在途预扣：在"冷启动无在途预扣"假设下，可用库存 = DB total。
// 该假设为既有限制（见 fix-inventory-cache-invalidation design D2），此处按可用库存口径写入。
func (s *ServiceContext) LoadInventoryFromDBToCache(ctx context.Context, productID int64) (*inventorymodel.Inventory, error) {
	inventoryRecord, err := s.InventoryModel.FindOne(ctx, productID)
	if err != nil {
		return nil, err
	}

	if err := s.SetInventoryCacheCtx(ctx, productID, inventoryRecord.Total); err != nil {
		return nil, err
	}

	return inventoryRecord, nil
}

func (s *ServiceContext) EnsureInventoryCacheCtx(ctx context.Context, productID int64) (int64, error) {
	total, ok, err := s.GetInventoryCacheCtx(ctx, productID)
	if err != nil {
		return 0, err
	}
	if ok {
		return total, nil
	}

	inventoryRecord, err := s.LoadInventoryFromDBToCache(ctx, productID)
	if err != nil {
		return 0, err
	}

	return inventoryRecord.Total, nil
}

func (s *ServiceContext) AdjustInventoryCacheCtx(ctx context.Context, productID, delta int64) (int64, error) {
	if _, err := s.EnsureInventoryCacheCtx(ctx, productID); err != nil {
		return 0, err
	}

	return s.Rdb.IncrbyCtx(ctx, inventoryCacheKey(productID), delta)
}
