package logic

import (
	"context"

	"services/internal/svc"
	"services/product/product/product/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAllProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllProductLogic {
	return &GetAllProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页得到全部商品
func (l *GetAllProductLogic) GetAllProduct(in *product.GetAllProductsReq) (*product.GetAllProductsResp, error) {
	// todo: add your logic here and delete this line

	return &product.GetAllProductsResp{}, nil
}
