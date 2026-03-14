package logic

import (
	"context"
	"errors"
	product2 "github.com/falconfan123/Go-mall/dal/model/products/product"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/product/internal/svc"
	product "github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAllProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllProductLogic {
	return &GetAllProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetAllProduct 分页得到全部商品
func (l *GetAllProductLogic) GetAllProduct(in *product.GetAllProductsReq) (*product.GetAllProductsResp, error) {

	// 并发查询数据
	var wg sync.WaitGroup
	var products []*product2.Products
	var total int64
	var queryErr error
	productModel := product2.NewProductsModel(l.svcCtx.Postgres)
	wg.Add(2)
	// 查询商品列表
	go func() {
		defer wg.Done()
		offset := (in.Page - 1) * in.PageSize
		products, queryErr = productModel.FindPage(l.ctx, int(offset), int(in.PageSize))
	}()

	// 查询总数
	go func() {
		defer wg.Done()
		total, queryErr = productModel.Count(l.ctx)
	}()
	wg.Wait()

	// 统一错误处理
	if queryErr != nil {
		if errors.Is(queryErr, sqlx.ErrNotFound) {
			// 也可以记录info日志
			// 返回空，可能是由于用户的过滤条件导致没有匹配到数据

			return &product.GetAllProductsResp{}, nil
		}
		l.Logger.Errorw("query products failed",
			logx.Field("err", queryErr),
			logx.Field("page", in.Page),
			logx.Field("pageSize", in.PageSize))
		return nil, queryErr
	}

	// Convert products to response format - without async processing first
	result := make([]*product.Product, len(products))
	for i, p := range products {
		result[i] = &product.Product{
			Id:          uint32(p.Id),
			Name:        p.Name,
			Description: p.Description.String,
			Picture:     p.Picture.String,
			Price:       p.Price,
			CratedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// Try to get inventory info (sync)
	for i, p := range products {
		if result[i] == nil {
			continue
		}
		// 默认使用商品表中的库存
		result[i].Stock = p.Stock

		inventoryResp, err := l.svcCtx.InventoryRpc.GetInventory(l.ctx, &inventoryclient.GetInventoryReq{
			ProductId: int32(p.Id),
		})
		if err != nil {
			logx.WithContext(l.ctx).Errorw("call InventoryRpc failed, use stock from products table", logx.Field("err", err), logx.Field("product_id", p.Id))
			// 使用商品表中的库存作为后备
			continue
		}
		// 如果 inventory 服务返回有效数据，使用它覆盖
		if inventoryResp.Inventory > 0 {
			result[i].Stock = inventoryResp.Inventory
		}
		if inventoryResp.SoldCount > 0 {
			result[i].Sold = inventoryResp.SoldCount
		}
	}

	return &product.GetAllProductsResp{
		Products: result,
		Total:    total,
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil
}
