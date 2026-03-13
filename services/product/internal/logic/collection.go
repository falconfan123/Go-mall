package logic

import (
	"context"
	product2 "github.com/falconfan123/Go-mall/dal/model/products/product"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/product/internal/svc"
	product "github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"sync"
	"time"
)

// populateProductDetails populates product details including inventory and categories
func populateProductDetails(ctx context.Context, svcCtx *svc.ServiceContext, products []*product2.Products) (result []*product.Product) {
	if len(products) == 0 {
		return make([]*product.Product, 0)
	}

	result = make([]*product.Product, len(products))

	// Populate basic product info
	for i, p := range products {
		result[i] = &product.Product{
			Id:          uint32(p.Id),
			Name:        p.Name,
			Description: p.Description.String,
			Picture:     p.Picture.String,
			Price:       p.Price,
			CratedAt:    p.CreatedAt.Format(time.DateTime),
			UpdatedAt:   p.UpdatedAt.Format(time.DateTime),
		}
	}

	// Process inventory and categories concurrently
	var wg sync.WaitGroup
	wg.Add(len(products) * 2)

	for i, p := range products {
		index := i
		productID := p.Id

		// Handle inventory
		wg.Add(1)
		go func() {
			defer wg.Done()
			handleInventory(ctx, svcCtx, result, index, productID)
		}()

		// Handle categories
		wg.Add(1)
		go func() {
			defer wg.Done()
			handleCategories(ctx, svcCtx, result, index, productID)
		}()
	}

	wg.Wait()
	return result
}

// 库存处理逻辑
func handleInventory(ctx context.Context, svcCtx *svc.ServiceContext, result []*product.Product, index int, productId int64) {
	inventoryResp, err := svcCtx.InventoryRpc.GetInventory(ctx, &inventoryclient.GetInventoryReq{
		ProductId: int32(productId),
	})
	if err != nil {
		logx.WithContext(ctx).Errorw("call InventoryRpc failed", logx.Field("err", err), logx.Field("product_id", productId))
		return
	}
	result[index].Stock = inventoryResp.Inventory
	result[index].Sold = inventoryResp.SoldCount
}

// 分类处理逻辑
func handleCategories(ctx context.Context, svcCtx *svc.ServiceContext, result []*product.Product, index int, productId int64) {
	categories, err := svcCtx.CategoriesModel.FindCategoryNameByProductID(ctx, productId)
	if err != nil {
		logx.WithContext(ctx).Errorw("query categories failed", logx.Field("err", err), logx.Field("product_id", productId))
		return
	}
	result[index].Categories = categories
}
