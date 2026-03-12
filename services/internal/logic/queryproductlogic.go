package logic

import (
	"context"

	"services/internal/svc"
	"services/product/product/product/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductLogic {
	return &QueryProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据条件查询商品
func (l *QueryProductLogic) QueryProduct(in *product.QueryProductReq) (*product.GetAllProductsResp, error) {
	// todo: add your logic here and delete this line

	return &product.GetAllProductsResp{}, nil
}
