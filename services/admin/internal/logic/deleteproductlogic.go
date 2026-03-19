package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
	product "github.com/falconfan123/Go-mall/services/product/pb"
)

type DeleteProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteProductLogic) DeleteProduct(in *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	client := product.NewProductCatalogServiceClient(l.svcCtx.ProductRpc.Conn())
	resp, err := client.DeleteProduct(l.ctx, &product.DeleteProductReq{Id: in.Id})
	if err != nil {
		return &pb.DeleteProductResponse{
			StatusCode: 500,
			StatusMsg:  "failed to delete product: " + err.Error(),
		}, nil
	}

	return &pb.DeleteProductResponse{
		StatusCode: resp.StatusCode,
		StatusMsg:  resp.StatusMsg,
	}, nil
}
