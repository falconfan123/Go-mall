package logic

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/services/admin/internal/db"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteActivityLogic {
	return &DeleteActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteActivityLogic) DeleteActivity(in *pb.DeleteActivityRequest) (*pb.DeleteActivityResponse, error) {
	activity, err := db.GetActivityByID(l.svcCtx.DB, in.Id)
	if err != nil {
		return &pb.DeleteActivityResponse{
			StatusCode: 404,
			StatusMsg:  "activity not found",
		}, nil
	}

	if err := activity.Delete(l.svcCtx.DB); err != nil {
		logx.Errorf("failed to delete activity: %v", err)
		return &pb.DeleteActivityResponse{
			StatusCode: 500,
			StatusMsg:  "failed to delete activity: " + err.Error(),
		}, nil
	}

	// Remove from Redis
	startKey := fmt.Sprintf("act_start_%d", in.Id)
	stockKey := fmt.Sprintf("act_%d_stock", in.Id)
	l.svcCtx.Redis.Del(startKey)
	l.svcCtx.Redis.Del(stockKey)

	return &pb.DeleteActivityResponse{
		StatusCode: 200,
		StatusMsg:  "success",
	}, nil
}
