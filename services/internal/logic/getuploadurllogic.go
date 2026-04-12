package logic

import (
	"context"

	"services/internal/svc"
	"services/product/product/product/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUploadURLLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUploadURLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUploadURLLogic {
	return &GetUploadURLLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MinIO 预签名上传接口
func (l *GetUploadURLLogic) GetUploadURL(in *product.GetUploadURLReq) (*product.GetUploadURLResp, error) {
	// todo: add your logic here and delete this line

	return &product.GetUploadURLResp{}, nil
}
