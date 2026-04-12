package server

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/logic"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	adminpb "github.com/falconfan123/Go-mall/services/admin/pb"
)

type AdminProductServiceServer struct {
	svcCtx *svc.ServiceContext
	adminpb.UnimplementedAdminProductServiceServer
}

func NewAdminProductServiceServer(svcCtx *svc.ServiceContext) *AdminProductServiceServer {
	return &AdminProductServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *AdminProductServiceServer) CreateProduct(ctx context.Context, in *adminpb.CreateProductRequest) (*adminpb.CreateProductResponse, error) {
	l := logic.NewCreateProductLogic(ctx, s.svcCtx)
	return l.CreateProduct(in)
}

func (s *AdminProductServiceServer) UpdateProduct(ctx context.Context, in *adminpb.UpdateProductRequest) (*adminpb.UpdateProductResponse, error) {
	l := logic.NewUpdateProductLogic(ctx, s.svcCtx)
	return l.UpdateProduct(in)
}

func (s *AdminProductServiceServer) DeleteProduct(ctx context.Context, in *adminpb.DeleteProductRequest) (*adminpb.DeleteProductResponse, error) {
	l := logic.NewDeleteProductLogic(ctx, s.svcCtx)
	return l.DeleteProduct(in)
}

func (s *AdminProductServiceServer) GetProduct(ctx context.Context, in *adminpb.GetProductRequest) (*adminpb.GetProductResponse, error) {
	l := logic.NewGetProductLogic(ctx, s.svcCtx)
	return l.GetProduct(in)
}

func (s *AdminProductServiceServer) ListProducts(ctx context.Context, in *adminpb.ListProductsRequest) (*adminpb.ListProductsResponse, error) {
	l := logic.NewListProductsLogic(ctx, s.svcCtx)
	return l.ListProducts(in)
}
