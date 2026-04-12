// Code generated manually for product client
package productclient

import (
	"context"

	"github.com/falconfan123/Go-mall/services/product/pb"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	GetProductReq       = pb.GetProductReq
	GetProductResp      = pb.GetProductResp
	GetAllProductsReq   = pb.GetAllProductsReq
	GetAllProductsResp  = pb.GetAllProductsResp
	CreateProductReq    = pb.CreateProductReq
	CreateProductResp   = pb.CreateProductResp
	UpdateProductReq    = pb.UpdateProductReq
	UpdateProductResp   = pb.UpdateProductResp
	DeleteProductReq    = pb.DeleteProductReq
	DeleteProductResp   = pb.DeleteProductResp
	IsExistProductReq   = pb.IsExistProductReq
	IsExistProductResp  = pb.IsExistProductResp
	QueryProductReq     = pb.QueryProductReq
	ListProductsReq     = pb.ListProductsReq
	ListProductsResp    = pb.ListProductsResp
	RecommendProductReq = pb.RecommendProductReq
	GetUploadURLReq     = pb.GetUploadURLReq
	GetUploadURLResp    = pb.GetUploadURLResp
	Product             = pb.Product

	ProductCatalog interface {
		GetProduct(ctx context.Context, in *GetProductReq, opts ...grpc.CallOption) (*GetProductResp, error)
		CreateProduct(ctx context.Context, in *CreateProductReq, opts ...grpc.CallOption) (*CreateProductResp, error)
		UpdateProduct(ctx context.Context, in *UpdateProductReq, opts ...grpc.CallOption) (*UpdateProductResp, error)
		DeleteProduct(ctx context.Context, in *DeleteProductReq, opts ...grpc.CallOption) (*DeleteProductResp, error)
		GetAllProduct(ctx context.Context, in *GetAllProductsReq, opts ...grpc.CallOption) (*GetAllProductsResp, error)
		IsExistProduct(ctx context.Context, in *IsExistProductReq, opts ...grpc.CallOption) (*IsExistProductResp, error)
		QueryProduct(ctx context.Context, in *QueryProductReq, opts ...grpc.CallOption) (*GetAllProductsResp, error)
		RecommendProduct(ctx context.Context, in *RecommendProductReq, opts ...grpc.CallOption) (*GetAllProductsResp, error)
		GetUploadURL(ctx context.Context, in *GetUploadURLReq, opts ...grpc.CallOption) (*GetUploadURLResp, error)
		ListProducts(ctx context.Context, in *ListProductsReq, opts ...grpc.CallOption) (*ListProductsResp, error)
	}

	defaultProductCatalog struct {
		cli zrpc.Client
	}
)

func NewProductCatalog(cli zrpc.Client) ProductCatalog {
	return &defaultProductCatalog{
		cli: cli,
	}
}

func (m *defaultProductCatalog) GetProduct(ctx context.Context, in *GetProductReq, opts ...grpc.CallOption) (*GetProductResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.GetProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) CreateProduct(ctx context.Context, in *CreateProductReq, opts ...grpc.CallOption) (*CreateProductResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.CreateProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) UpdateProduct(ctx context.Context, in *UpdateProductReq, opts ...grpc.CallOption) (*UpdateProductResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.UpdateProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) DeleteProduct(ctx context.Context, in *DeleteProductReq, opts ...grpc.CallOption) (*DeleteProductResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.DeleteProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) GetAllProduct(ctx context.Context, in *GetAllProductsReq, opts ...grpc.CallOption) (*GetAllProductsResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.GetAllProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) IsExistProduct(ctx context.Context, in *IsExistProductReq, opts ...grpc.CallOption) (*IsExistProductResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.IsExistProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) QueryProduct(ctx context.Context, in *QueryProductReq, opts ...grpc.CallOption) (*GetAllProductsResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.QueryProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) RecommendProduct(ctx context.Context, in *RecommendProductReq, opts ...grpc.CallOption) (*GetAllProductsResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.RecommendProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) GetUploadURL(ctx context.Context, in *GetUploadURLReq, opts ...grpc.CallOption) (*GetUploadURLResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.GetUploadURL(ctx, in, opts...)
}

func (m *defaultProductCatalog) ListProducts(ctx context.Context, in *ListProductsReq, opts ...grpc.CallOption) (*ListProductsResp, error) {
	client := pb.NewProductCatalogServiceClient(m.cli.Conn())
	return client.ListProducts(ctx, in, opts...)
}
