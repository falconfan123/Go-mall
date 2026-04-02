package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	adminpb "github.com/falconfan123/Go-mall/services/admin/pb"
)

type DeleteCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCategoryLogic) DeleteCategory(in *adminpb.DeleteCategoryRequest) (*adminpb.DeleteCategoryResponse, error) {
	return &adminpb.DeleteCategoryResponse{
		StatusCode: 200,
		StatusMsg:  "success",
	}, nil
}
