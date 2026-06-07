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
