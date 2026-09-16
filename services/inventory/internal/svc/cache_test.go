package svc

import (
	"context"
	"testing"

	inventorymodel "github.com/falconfan123/Go-mall/dal/model/inventory"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// mockInventoryModel 仅实现本测试用到的 InventoryModel 方法。
type mockInventoryModel struct {
	inventorymodel.InventoryModel
	findAllFn func(ctx context.Context) ([]*inventorymodel.Inventory, error)
	findOneFn func(ctx context.Context, productId int64) (*inventorymodel.Inventory, error)
}

func (m *mockInventoryModel) FindAll(ctx context.Context) ([]*inventorymodel.Inventory, error) {
	if m.findAllFn == nil {
		return nil, nil
	}
	return m.findAllFn(ctx)
}

func (m *mockInventoryModel) FindOne(ctx context.Context, productId int64) (*inventorymodel.Inventory, error) {
	if m.findOneFn == nil {
		return nil, inventorymodel.ErrNotFound
	}
	return m.findOneFn(ctx, productId)
}

func newTestSvcCtx(m *mockInventoryModel) *ServiceContext {
	return &ServiceContext{
		Rdb:            redis.MustNewRedis(redis.RedisConf{Host: "localhost:6379", Type: "node"}),
		InventoryModel: m,
	}
}

func TestLoadInventoryFromDBToCache_WritesAvailable(t *testing.T) {
	const pid = int64(888801)
	ctx := context.Background()
	m := &mockInventoryModel{
		findOneFn: func(_ context.Context, _ int64) (*inventorymodel.Inventory, error) {
			return &inventorymodel.Inventory{ProductId: pid, Total: 100, Sold: 0}, nil
		},
	}
	svcCtx := newTestSvcCtx(m)
	_, _ = svcCtx.Rdb.DelCtx(ctx, inventoryCacheKey(pid))
	defer svcCtx.Rdb.DelCtx(ctx, inventoryCacheKey(pid))

	rec, err := svcCtx.LoadInventoryFromDBToCache(ctx, pid)
	if err != nil {
		t.Fatalf("LoadInventoryFromDBToCache err: %v", err)
	}
	if rec.Total != 100 {
		t.Fatalf("record total = %d, want 100", rec.Total)
	}

	got, ok, err := svcCtx.GetInventoryCacheCtx(ctx, pid)
	if err != nil {
		t.Fatalf("GetInventoryCacheCtx err: %v", err)
	}
	if !ok {
		t.Fatal("cache key not set after backfill")
	}
	if got != 100 {
		t.Fatalf("cached available = %d, want 100 (cold-start 预留=0 → available=total)", got)
	}
}

func TestPreheatInventoryCache_OnlyBackfillsMissingKeys(t *testing.T) {
	const (
		pidExisting = int64(888802) // 已存在键，不应被覆盖
		pidMissing  = int64(888803) // 缺失键，应回填 total
	)
	ctx := context.Background()
	m := &mockInventoryModel{
		findAllFn: func(_ context.Context) ([]*inventorymodel.Inventory, error) {
			return []*inventorymodel.Inventory{
				{ProductId: pidExisting, Total: 100, Sold: 0},
				{ProductId: pidMissing, Total: 200, Sold: 0},
			}, nil
		},
	}
	svcCtx := newTestSvcCtx(m)
	_, _ = svcCtx.Rdb.DelCtx(ctx, inventoryCacheKey(pidExisting))
	_, _ = svcCtx.Rdb.DelCtx(ctx, inventoryCacheKey(pidMissing))
	defer func() {
		svcCtx.Rdb.DelCtx(ctx, inventoryCacheKey(pidExisting))
		svcCtx.Rdb.DelCtx(ctx, inventoryCacheKey(pidMissing))
	}()

	// 预置已存在键为一个非 total 的可用值（模拟在途预扣后的可用库存）
	if err := svcCtx.Rdb.SetCtx(ctx, inventoryCacheKey(pidExisting), "42"); err != nil {
		t.Fatalf("preset existing key err: %v", err)
	}

	if err := svcCtx.PreheatInventoryCache(); err != nil {
		t.Fatalf("PreheatInventoryCache err: %v", err)
	}

	existing, ok, err := svcCtx.GetInventoryCacheCtx(ctx, pidExisting)
	if err != nil {
		t.Fatalf("read existing key err: %v", err)
	}
	if !ok {
		t.Fatal("existing key missing after preheat")
	}
	if existing != 42 {
		t.Fatalf("existing key overwritten: got %d, want 42 (MUST NOT overwrite in-flight available)", existing)
	}

	missing, ok, err := svcCtx.GetInventoryCacheCtx(ctx, pidMissing)
	if err != nil {
		t.Fatalf("read missing key err: %v", err)
	}
	if !ok {
		t.Fatal("missing key not backfilled")
	}
	if missing != 200 {
		t.Fatalf("missing key value = %d, want 200 (backfill available=total)", missing)
	}
}
