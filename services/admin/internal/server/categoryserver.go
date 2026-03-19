package server

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/logic"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
)

type AdminCategoryServiceServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedAdminCategoryServiceServer
}

func NewAdminCategoryServiceServer(svcCtx *svc.ServiceContext) *AdminCategoryServiceServer {
	return &AdminCategoryServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *AdminCategoryServiceServer) CreateCategory(ctx context.Context, in *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	l := logic.NewCreateCategoryLogic(ctx, s.svcCtx)
	return l.CreateCategory(in)
}

func (s *AdminCategoryServiceServer) UpdateCategory(ctx context.Context, in *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	l := logic.NewUpdateCategoryLogic(ctx, s.svcCtx)
	return l.UpdateCategory(in)
}

func (s *AdminCategoryServiceServer) DeleteCategory(ctx context.Context, in *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	l := logic.NewDeleteCategoryLogic(ctx, s.svcCtx)
	return l.DeleteCategory(in)
}

func (s *AdminCategoryServiceServer) ListCategories(ctx context.Context, in *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	l := logic.NewListCategoriesLogic(ctx, s.svcCtx)
	return l.ListCategories(in)
}
