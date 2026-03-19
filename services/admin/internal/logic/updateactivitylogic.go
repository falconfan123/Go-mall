package logic

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/services/admin/internal/db"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateActivityLogic {
	return &UpdateActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateActivityLogic) UpdateActivity(in *pb.UpdateActivityRequest) (*pb.UpdateActivityResponse, error) {
	activity, err := db.GetActivityByID(l.svcCtx.DB, in.Id)
	if err != nil {
		return &pb.UpdateActivityResponse{
			StatusCode: 404,
			StatusMsg:  "activity not found",
		}, nil
	}

	activity.Name = in.Name
	activity.SeckillPrice = in.SeckillPrice
	activity.TotalStock = in.TotalStock
	activity.LimitPerUser = in.LimitPerUser

	if err := activity.Update(l.svcCtx.DB); err != nil {
		logx.Errorf("failed to update activity: %v", err)
		return &pb.UpdateActivityResponse{
			StatusCode: 500,
			StatusMsg:  "failed to update activity: " + err.Error(),
		}, nil
	}

	// Sync to Redis
	l.syncActivityToRedis(activity)

	return &pb.UpdateActivityResponse{
		StatusCode: 200,
		StatusMsg:  "success",
		Id:         in.Id,
	}, nil
}

func (l *UpdateActivityLogic) syncActivityToRedis(activity *db.Activity) {
	stockKey := fmt.Sprintf("act_%d_stock", activity.ID)
	l.svcCtx.Redis.Set(stockKey, fmt.Sprintf("%d", activity.TotalStock))

	logx.Infof("updated activity %d stock in Redis: stock=%d",
		activity.ID, activity.TotalStock)
}
