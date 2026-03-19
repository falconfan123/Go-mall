package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
	product "github.com/falconfan123/Go-mall/services/product/pb"
)

type CreateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateProductLogic) CreateProduct(in *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	// Call product service RPC to create product
	client := product.NewProductCatalogServiceClient(l.svcCtx.ProductRpc.Conn())
	resp, err := client.CreateProduct(l.ctx, &product.CreateProductReq{
		Name:        in.Name,
		Description: in.Description,
		Picture:     in.Picture,
		Price:       in.Price,
		Stock:       in.Stock,
		Categories:  in.Categories,
	})
	if err != nil {
		return &pb.CreateProductResponse{
			StatusCode: 500,
			StatusMsg:  "failed to create product: " + err.Error(),
		}, nil
	}

	return &pb.CreateProductResponse{
		StatusCode: resp.StatusCode,
		StatusMsg:  resp.StatusMsg,
		ProductId:  resp.ProductId,
	}, nil
}
