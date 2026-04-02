package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	adminpb "github.com/falconfan123/Go-mall/services/admin/pb"
	product "github.com/falconfan123/Go-mall/services/product/pb"
)

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProductsLogic) ListProducts(in *adminpb.ListProductsRequest) (*adminpb.ListProductsResponse, error) {
	client := product.NewProductCatalogServiceClient(l.svcCtx.ProductRpc.Conn())
	resp, err := client.GetAllProduct(l.ctx, &product.GetAllProductsReq{
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		return &adminpb.ListProductsResponse{
			StatusCode: 500,
			StatusMsg:  "failed to list products: " + err.Error(),
		}, nil
	}

	products := make([]*adminpb.Product, len(resp.Products))
	for i, p := range resp.Products {
		products[i] = convertProduct(p)
	}

	return &adminpb.ListProductsResponse{
		StatusCode: resp.StatusCode,
		StatusMsg:  resp.StatusMsg,
		Total:      resp.Total,
		Products:   products,
	}, nil
}
