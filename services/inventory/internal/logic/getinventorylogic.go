package logic

import (
	"context"
	"errors"
	"github.com/falconfan123/Go-mall/common/consts/code"
	inventorymodel "github.com/falconfan123/Go-mall/dal/model/inventory"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/falconfan123/Go-mall/services/inventory/internal/svc"
	inventory "github.com/falconfan123/Go-mall/services/inventory/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInventoryLogic {
	return &GetInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetInventory 查询库存
func (l *GetInventoryLogic) GetInventory(in *inventory.GetInventoryReq) (*inventory.GetInventoryResp, error) {
	res := new(inventory.GetInventoryResp)

	inventoryResp, err := l.svcCtx.InventoryModel.FindOne(l.ctx, int64(in.ProductId))
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.Logger.Infow("product not in inventory", logx.Field("product_id", in.ProductId))
			res.StatusCode = code.ProductNotFoundInventory
			res.StatusMsg = code.ProductNotFoundInventoryMsg
			return res, nil
		}
		l.Logger.Errorw("product inventory get failed", logx.Field("product_id", in.ProductId))
		return nil, err
	}

	cachedTotal, cached, err := l.svcCtx.GetInventoryCacheCtx(l.ctx, int64(in.ProductId))
	if err != nil {
		l.Logger.Errorw("read inventory cache failed",
			logx.Field("err", err),
			logx.Field("product_id", in.ProductId),
		)
		return nil, err
	}
	if !cached {
		if _, err := l.svcCtx.LoadInventoryFromDBToCache(l.ctx, int64(in.ProductId)); err != nil {
			if errors.Is(err, inventorymodel.ErrNotFound) {
				l.Logger.Infow("product not in inventory", logx.Field("product_id", in.ProductId))
				res.StatusCode = code.ProductNotFoundInventory
				res.StatusMsg = code.ProductNotFoundInventoryMsg
				return res, nil
			}
			l.Logger.Errorw("backfill inventory cache failed",
				logx.Field("err", err),
				logx.Field("product_id", in.ProductId),
			)
			return nil, err
		}
		cachedTotal = inventoryResp.Total
	}

	res.Inventory = cachedTotal
	res.SoldCount = inventoryResp.Sold
	return res, nil
}
