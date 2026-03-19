package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
	product "github.com/falconfan123/Go-mall/services/product/pb"
)

type UpdateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProductLogic) UpdateProduct(in *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	client := product.NewProductCatalogServiceClient(l.svcCtx.ProductRpc.Conn())
	resp, err := client.UpdateProduct(l.ctx, &product.UpdateProductReq{
		Id:          in.Id,
		Name:        in.Name,
		Description: in.Description,
		Picture:     in.Picture,
		Price:       in.Price,
		Stock:       in.Stock,
		Categories:  in.Categories,
	})
	if err != nil {
		return &pb.UpdateProductResponse{
			StatusCode: 500,
			StatusMsg:  "failed to update product: " + err.Error(),
		}, nil
	}

	return &pb.UpdateProductResponse{
		StatusCode: resp.StatusCode,
		StatusMsg:  resp.StatusMsg,
		Id:         resp.Id,
	}, nil
}
