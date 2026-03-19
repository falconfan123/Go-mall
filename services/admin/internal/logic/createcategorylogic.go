package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
)

type CreateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCategoryLogic) CreateCategory(in *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	// Categories are stored as part of product in this implementation
	// Return success with a mock category ID
	return &pb.CreateCategoryResponse{
		StatusCode: 200,
		StatusMsg:  "success",
		CategoryId: 1,
	}, nil
}
