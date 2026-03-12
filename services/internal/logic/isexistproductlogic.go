package logic

import (
	"context"

	"services/internal/svc"
	"services/product/product/product/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsExistProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsExistProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsExistProductLogic {
	return &IsExistProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 判断商品是否存在
func (l *IsExistProductLogic) IsExistProduct(in *product.IsExistProductReq) (*product.IsExistProductResp, error) {
	// todo: add your logic here and delete this line

	return &product.IsExistProductResp{}, nil
}
