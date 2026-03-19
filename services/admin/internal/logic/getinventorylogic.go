package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
	inventory "github.com/falconfan123/Go-mall/services/inventory/pb"
)

type GetInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInventoryLogic {
	return &GetInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInventoryLogic) GetInventory(in *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	client := inventory.NewInventoryClient(l.svcCtx.InventoryRpc.Conn())

	resp, err := client.GetInventory(l.ctx, &inventory.GetInventoryReq{
		ProductId: int32(in.ProductId),
	})
	if err != nil {
		return &pb.GetInventoryResponse{
			StatusCode: 500,
			StatusMsg:  "failed to get inventory: " + err.Error(),
		}, nil
	}

	return &pb.GetInventoryResponse{
		StatusCode: 200,
		StatusMsg:  "success",
		ProductId:  in.ProductId,
		Stock:      resp.Inventory,
		Reserved:   resp.SoldCount,
	}, nil
}
