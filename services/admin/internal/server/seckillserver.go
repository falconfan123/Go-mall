package server

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/logic"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
)

type AdminSeckillServiceServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedAdminSeckillServiceServer
}

func NewAdminSeckillServiceServer(svcCtx *svc.ServiceContext) *AdminSeckillServiceServer {
	return &AdminSeckillServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *AdminSeckillServiceServer) CreateActivity(ctx context.Context, in *pb.CreateActivityRequest) (*pb.CreateActivityResponse, error) {
	l := logic.NewCreateActivityLogic(ctx, s.svcCtx)
	return l.CreateActivity(in)
}

func (s *AdminSeckillServiceServer) UpdateActivity(ctx context.Context, in *pb.UpdateActivityRequest) (*pb.UpdateActivityResponse, error) {
	l := logic.NewUpdateActivityLogic(ctx, s.svcCtx)
	return l.UpdateActivity(in)
}

func (s *AdminSeckillServiceServer) DeleteActivity(ctx context.Context, in *pb.DeleteActivityRequest) (*pb.DeleteActivityResponse, error) {
	l := logic.NewDeleteActivityLogic(ctx, s.svcCtx)
	return l.DeleteActivity(in)
}

func (s *AdminSeckillServiceServer) GetActivity(ctx context.Context, in *pb.GetActivityRequest) (*pb.GetActivityResponse, error) {
	l := logic.NewGetActivityLogic(ctx, s.svcCtx)
	return l.GetActivity(in)
}

func (s *AdminSeckillServiceServer) ListActivities(ctx context.Context, in *pb.ListActivitiesRequest) (*pb.ListActivitiesResponse, error) {
	l := logic.NewListActivitiesLogic(ctx, s.svcCtx)
	return l.ListActivities(in)
}

func (s *AdminSeckillServiceServer) SetActivityTime(ctx context.Context, in *pb.SetActivityTimeRequest) (*pb.SetActivityTimeResponse, error) {
	l := logic.NewSetActivityTimeLogic(ctx, s.svcCtx)
	return l.SetActivityTime(in)
}

func (s *AdminSeckillServiceServer) SetActivityStock(ctx context.Context, in *pb.SetActivityStockRequest) (*pb.SetActivityStockResponse, error) {
	l := logic.NewSetActivityStockLogic(ctx, s.svcCtx)
	return l.SetActivityStock(in)
}
