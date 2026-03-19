package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/db"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
)

type ListActivitiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListActivitiesLogic {
	return &ListActivitiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListActivitiesLogic) ListActivities(in *pb.ListActivitiesRequest) (*pb.ListActivitiesResponse, error) {
	activities, total, err := db.ListActivities(l.svcCtx.DB, in.Page, in.PageSize, in.Status)
	if err != nil {
		return &pb.ListActivitiesResponse{
			StatusCode: 500,
			StatusMsg:  "failed to list activities: " + err.Error(),
		}, nil
	}

	pbActivities := make([]*pb.Activity, len(activities))
	for i, a := range activities {
		pbActivities[i] = convertActivity(a)
	}

	return &pb.ListActivitiesResponse{
		StatusCode: 200,
		StatusMsg:  "success",
		Total:      total,
		Activities: pbActivities,
	}, nil
}
