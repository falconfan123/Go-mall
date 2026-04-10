package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	adminpb "github.com/falconfan123/Go-mall/services/admin/pb"
	product "github.com/falconfan123/Go-mall/services/product/pb"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductLogic) GetProduct(in *adminpb.GetProductRequest) (*adminpb.GetProductResponse, error) {
	client := product.NewProductCatalogServiceClient(l.svcCtx.ProductRpc.Conn())
	resp, err := client.GetProduct(l.ctx, &product.GetProductReq{Id: uint32(in.Id)})
	if err != nil {
		return &adminpb.GetProductResponse{
			StatusCode: 500,
			StatusMsg:  "failed to get product: " + err.Error(),
		}, nil
	}

	return &adminpb.GetProductResponse{
		StatusCode: resp.StatusCode,
		StatusMsg:  resp.StatusMsg,
		Product:    convertProduct(resp.Product),
	}, nil
}

func convertProduct(p *product.Product) *adminpb.Product {
	if p == nil {
		return nil
	}
	return &adminpb.Product{
		Id:          p.Id,
		Name:        p.Name,
		Description: p.Description,
		Picture:     p.Picture,
		Price:       p.Price,
		Stock:       p.Stock,
		Sold:        p.Sold,
		Categories:  p.Categories,
		CreatedAt:   p.CratedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
