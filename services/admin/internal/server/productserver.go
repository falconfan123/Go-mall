package server

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/logic"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
)

type AdminProductServiceServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedAdminProductServiceServer
}

func NewAdminProductServiceServer(svcCtx *svc.ServiceContext) *AdminProductServiceServer {
	return &AdminProductServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *AdminProductServiceServer) CreateProduct(ctx context.Context, in *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	l := logic.NewCreateProductLogic(ctx, s.svcCtx)
	return l.CreateProduct(in)
}

func (s *AdminProductServiceServer) UpdateProduct(ctx context.Context, in *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	l := logic.NewUpdateProductLogic(ctx, s.svcCtx)
	return l.UpdateProduct(in)
}

func (s *AdminProductServiceServer) DeleteProduct(ctx context.Context, in *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	l := logic.NewDeleteProductLogic(ctx, s.svcCtx)
	return l.DeleteProduct(in)
}

func (s *AdminProductServiceServer) GetProduct(ctx context.Context, in *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	l := logic.NewGetProductLogic(ctx, s.svcCtx)
	return l.GetProduct(in)
}

func (s *AdminProductServiceServer) ListProducts(ctx context.Context, in *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	l := logic.NewListProductsLogic(ctx, s.svcCtx)
	return l.ListProducts(in)
}
