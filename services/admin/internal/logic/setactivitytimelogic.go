package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/services/admin/internal/db"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	adminpb "github.com/falconfan123/Go-mall/services/admin/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type SetActivityTimeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetActivityTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetActivityTimeLogic {
	return &SetActivityTimeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetActivityTimeLogic) SetActivityTime(in *adminpb.SetActivityTimeRequest) (*adminpb.SetActivityTimeResponse, error) {
	activity, err := db.GetActivityByID(l.svcCtx.DB, in.Id)
	if err != nil {
		return &adminpb.SetActivityTimeResponse{
			StatusCode: 404,
			StatusMsg:  "activity not found",
		}, nil
	}

	startTime, err := time.Parse(time.RFC3339, in.StartTime)
	if err != nil {
		startTime, err = time.Parse("2006-01-02 15:04:05", in.StartTime)
		if err != nil {
			return &adminpb.SetActivityTimeResponse{
				StatusCode: 400,
				StatusMsg:  "invalid start time format",
			}, nil
		}
	}

	endTime, err := time.Parse(time.RFC3339, in.EndTime)
	if err != nil {
		endTime, err = time.Parse("2006-01-02 15:04:05", in.EndTime)
		if err != nil {
			return &adminpb.SetActivityTimeResponse{
				StatusCode: 400,
				StatusMsg:  "invalid end time format",
			}, nil
		}
	}

	activity.StartTime = startTime
	activity.EndTime = endTime

	// Update status based on time
	now := time.Now()
	if now.After(startTime) && now.Before(endTime) {
		activity.Status = "active"
	} else if now.After(endTime) {
		activity.Status = "ended"
	} else {
		activity.Status = "pending"
	}

	if err := activity.Update(l.svcCtx.DB); err != nil {
		logx.Errorf("failed to update activity time: %v", err)
		return &adminpb.SetActivityTimeResponse{
			StatusCode: 500,
			StatusMsg:  "failed to update activity time: " + err.Error(),
		}, nil
	}

	// Sync to Redis with TTL (活动结束后多保留1天)
	startKey := fmt.Sprintf("act_start_%d", activity.ID)
	expireSeconds := time.Until(activity.EndTime).Seconds()
	if expireSeconds < 0 {
		expireSeconds = biz.SeckillCacheTTL.Seconds()
	} else {
		expireSeconds += biz.SeckillCacheTTL.Seconds()
	}
	l.svcCtx.Redis.Setex(startKey, fmt.Sprintf("%d", activity.StartTime.UnixMilli()), int(expireSeconds))

	logx.Infof("updated activity %d time in Redis: start=%d, ttl=%d seconds",
		activity.ID, activity.StartTime.UnixMilli(), int(expireSeconds))

	return &adminpb.SetActivityTimeResponse{
		StatusCode: 200,
		StatusMsg:  "success",
	}, nil
}
