package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	adminpb "github.com/falconfan123/Go-mall/services/admin/pb"
)

type ListCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCategoriesLogic) ListCategories(in *adminpb.ListCategoriesRequest) (*adminpb.ListCategoriesResponse, error) {
	// Return default categories
	return &adminpb.ListCategoriesResponse{
		StatusCode: 200,
		StatusMsg:  "success",
		Categories: []*adminpb.Category{
			{Id: 1, Name: "电子产品", Description: "电子相关产品", ParentId: 0},
			{Id: 2, Name: "服装", Description: "服装类", ParentId: 0},
			{Id: 3, Name: "食品", Description: "食品类", ParentId: 0},
		},
	}, nil
}
