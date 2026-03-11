package repository

import (
	"context"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/aggregate"
)

// ProductRepository 商品仓储接口，定义对商品聚合根的操作
type ProductRepository interface {
	// Save 保存商品（新建或更新）
	Save(ctx context.Context, product *aggregate.Product) error

	// GetByID 根据ID查询商品
	GetByID(ctx context.Context, id int64) (*aggregate.Product, error)

	// List 查询商品列表
	List(ctx context.Context, page, pageSize int, categoryID *int64, keyword *string) ([]*aggregate.Product, int64, error)

	// Delete 删除商品
	Delete(ctx context.Context, id int64) error

	// BatchGetByIDs 批量查询商品
	BatchGetByIDs(ctx context.Context, ids []int64) ([]*aggregate.Product, error)

	// DecreaseStock 扣减库存（原子操作）
	DecreaseStock(ctx context.Context, productID int64, quantity int64) error

	// IncreaseStock 增加库存（原子操作）
	IncreaseStock(ctx context.Context, productID int64, quantity int64) error
}
