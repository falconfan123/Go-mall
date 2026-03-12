// Code generated manually for product client
package productclient

import (
	"context"

	"github.com/falconfan123/Go-mall/services/product/product"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	GetProductReq         = product.GetProductReq
	GetProductResp        = product.GetProductResp
	GetAllProductsReq     = product.GetAllProductsReq
	GetAllProductsResp    = product.GetAllProductsResp
	CreateProductReq     = product.CreateProductReq
	CreateProductResp    = product.CreateProductResp
	UpdateProductReq     = product.UpdateProductReq
	UpdateProductResp    = product.UpdateProductResp
	DeleteProductReq     = product.DeleteProductReq
	DeleteProductResp    = product.DeleteProductResp
	IsExistProductReq    = product.IsExistProductReq
	IsExistProductResp   = product.IsExistProductResp
	QueryProductReq      = product.QueryProductReq
	ListProductsReq      = product.ListProductsReq
	ListProductsResp     = product.ListProductsResp
	RecommendProductReq  = product.RecommendProductReq
	GetUploadURLReq     = product.GetUploadURLReq
	GetUploadURLResp    = product.GetUploadURLResp
	Product             = product.Product

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
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.GetProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) CreateProduct(ctx context.Context, in *CreateProductReq, opts ...grpc.CallOption) (*CreateProductResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.CreateProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) UpdateProduct(ctx context.Context, in *UpdateProductReq, opts ...grpc.CallOption) (*UpdateProductResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.UpdateProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) DeleteProduct(ctx context.Context, in *DeleteProductReq, opts ...grpc.CallOption) (*DeleteProductResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.DeleteProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) GetAllProduct(ctx context.Context, in *GetAllProductsReq, opts ...grpc.CallOption) (*GetAllProductsResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.GetAllProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) IsExistProduct(ctx context.Context, in *IsExistProductReq, opts ...grpc.CallOption) (*IsExistProductResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.IsExistProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) QueryProduct(ctx context.Context, in *QueryProductReq, opts ...grpc.CallOption) (*GetAllProductsResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.QueryProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) RecommendProduct(ctx context.Context, in *RecommendProductReq, opts ...grpc.CallOption) (*GetAllProductsResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.RecommendProduct(ctx, in, opts...)
}

func (m *defaultProductCatalog) GetUploadURL(ctx context.Context, in *GetUploadURLReq, opts ...grpc.CallOption) (*GetUploadURLResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.GetUploadURL(ctx, in, opts...)
}

func (m *defaultProductCatalog) ListProducts(ctx context.Context, in *ListProductsReq, opts ...grpc.CallOption) (*ListProductsResp, error) {
	client := product.NewProductCatalogServiceClient(m.cli.Conn())
	return client.ListProducts(ctx, in, opts...)
}
