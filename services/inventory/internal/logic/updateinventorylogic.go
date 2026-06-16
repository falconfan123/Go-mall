package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	inventory2 "github.com/falconfan123/Go-mall/dal/model/inventory"
	"github.com/falconfan123/Go-mall/services/inventory/internal/svc"
	inventory "github.com/falconfan123/Go-mall/services/inventory/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateInventoryLogic {
	return &UpdateInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateInventory 更新库存，进行修改库存数量
func (l *UpdateInventoryLogic) UpdateInventory(in *inventory.UpdateInventoryReq) (*inventory.InventoryResp, error) {

	for _, item := range in.Items {

		if item.Quantity <= 0 {
			l.Logger.Errorw("quantity must be greater than 0", logx.Field("quantity", item.Quantity), logx.Field("product_id", item.ProductId))
			return nil, biz.ErrInvalidInventory
		}
		//执行sql
		if err := l.svcCtx.InventoryModel.UpdateOrCreate(l.ctx, inventory2.Inventory{
			ProductId: int64(item.ProductId),
			Total:     int64(item.Quantity),
		}); err != nil {
			l.Logger.Errorw("update inventory error", logx.Field("error", err.Error()), logx.Field("product_id", item.ProductId))
			return nil, err
		}

		if err := l.svcCtx.SetInventoryCacheCtx(l.ctx, int64(item.ProductId), int64(item.Quantity)); err != nil {
			l.Logger.Errorw("update inventory cache failed",
				logx.Field("product_id", item.ProductId),
				logx.Field("err", err),
			)
		}
	}
	return &inventory.InventoryResp{}, nil

}
