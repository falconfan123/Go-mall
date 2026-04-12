package logic

import (
	"context"

	"services/internal/svc"
	"services/product/product/product/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecommendProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecommendProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecommendProductLogic {
	return &RecommendProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RecommendProductLogic) RecommendProduct(in *product.RecommendProductReq) (*product.GetAllProductsResp, error) {
	// todo: add your logic here and delete this line

	return &product.GetAllProductsResp{}, nil
}
