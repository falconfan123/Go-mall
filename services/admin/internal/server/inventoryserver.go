package server

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/logic"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
)

type AdminInventoryServiceServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedAdminInventoryServiceServer
}

func NewAdminInventoryServiceServer(svcCtx *svc.ServiceContext) *AdminInventoryServiceServer {
	return &AdminInventoryServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *AdminInventoryServiceServer) AdjustInventory(ctx context.Context, in *pb.AdjustInventoryRequest) (*pb.AdjustInventoryResponse, error) {
	l := logic.NewAdjustInventoryLogic(ctx, s.svcCtx)
	return l.AdjustInventory(in)
}

func (s *AdminInventoryServiceServer) GetInventory(ctx context.Context, in *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	l := logic.NewGetInventoryLogic(ctx, s.svcCtx)
	return l.GetInventory(in)
}
