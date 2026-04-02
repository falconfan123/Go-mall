package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	adminpb "github.com/falconfan123/Go-mall/services/admin/pb"
	inventory "github.com/falconfan123/Go-mall/services/inventory/pb"
)

type AdjustInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdjustInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdjustInventoryLogic {
	return &AdjustInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdjustInventoryLogic) AdjustInventory(in *adminpb.AdjustInventoryRequest) (*adminpb.AdjustInventoryResponse, error) {
	client := inventory.NewInventoryClient(l.svcCtx.InventoryRpc.Conn())

	// Get current inventory first
	getResp, err := client.GetInventory(l.ctx, &inventory.GetInventoryReq{
		ProductId: int32(in.ProductId),
	})
	if err != nil {
		return &adminpb.AdjustInventoryResponse{
			StatusCode: 500,
			StatusMsg:  "failed to get inventory: " + err.Error(),
		}, nil
	}

	// Calculate new inventory
	newStock := getResp.Inventory + in.Quantity
	if newStock < 0 {
		return &adminpb.AdjustInventoryResponse{
			StatusCode: 400,
			StatusMsg:  "insufficient inventory",
		}, nil
	}

	// Update inventory
	_, err = client.UpdateInventory(l.ctx, &inventory.UpdateInventoryReq{
		Items: []*inventory.UpdateInventoryReq_Items{
			{ProductId: int32(in.ProductId), Quantity: int32(in.Quantity)},
		},
	})
	if err != nil {
		return &adminpb.AdjustInventoryResponse{
			StatusCode: 500,
			StatusMsg:  "failed to adjust inventory: " + err.Error(),
		}, nil
	}

	return &adminpb.AdjustInventoryResponse{
		StatusCode: 200,
		StatusMsg:  "success",
		NewStock:   newStock,
	}, nil
}
