package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
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

func (l *DeleteCategoryLogic) DeleteCategory(in *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	return &pb.DeleteCategoryResponse{
		StatusCode: 200,
		StatusMsg:  "success",
	}, nil
}
