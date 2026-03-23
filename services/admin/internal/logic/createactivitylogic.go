package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/services/admin/internal/db"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateActivityLogic {
	return &CreateActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateActivityLogic) CreateActivity(in *pb.CreateActivityRequest) (*pb.CreateActivityResponse, error) {
	// Parse time
	startTime, err := time.Parse(time.RFC3339, in.StartTime)
	if err != nil {
		startTime, err = time.Parse("2006-01-02 15:04:05", in.StartTime)
		if err != nil {
			return &pb.CreateActivityResponse{
				StatusCode: 400,
				StatusMsg:  "invalid start time format",
			}, nil
		}
	}

	endTime, err := time.Parse(time.RFC3339, in.EndTime)
	if err != nil {
		endTime, err = time.Parse("2006-01-02 15:04:05", in.EndTime)
		if err != nil {
			return &pb.CreateActivityResponse{
				StatusCode: 400,
				StatusMsg:  "invalid end time format",
			}, nil
		}
	}

	// Determine status based on time
	status := "pending"
	now := time.Now()
	if now.After(startTime) && now.Before(endTime) {
		status = "active"
	}

	// Create activity in database
	activity := &db.Activity{
		Name:         in.Name,
		ProductID:    in.ProductId,
		SeckillPrice: in.SeckillPrice,
		TotalStock:   in.TotalStock,
		LimitPerUser: in.LimitPerUser,
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       status,
	}

	if err := activity.Create(l.svcCtx.DB); err != nil {
		logx.Errorf("failed to create activity: %v", err)
		return &pb.CreateActivityResponse{
			StatusCode: 500,
			StatusMsg:  "failed to create activity: " + err.Error(),
		}, nil
	}

	// Sync to Redis
	l.syncActivityToRedis(activity)

	return &pb.CreateActivityResponse{
		StatusCode: 200,
		StatusMsg:  "success",
		ActivityId: activity.ID,
	}, nil
}

func (l *CreateActivityLogic) syncActivityToRedis(activity *db.Activity) {
	// 计算活动结束后的过期时间，额外多保留一天
	expireSeconds := int(activity.EndTime.Sub(time.Now()).Seconds()) + int(biz.SeckillCacheTTL.Seconds())
	if expireSeconds < 0 {
		expireSeconds = int(biz.SeckillCacheTTL.Seconds())
	}

	// Set activity start time with TTL
	startKey := fmt.Sprintf("act_start_%d", activity.ID)
	l.svcCtx.Redis.Setex(startKey, fmt.Sprintf("%d", activity.StartTime.UnixMilli()), expireSeconds)

	// Set activity stock with TTL
	stockKey := fmt.Sprintf("act_%d_stock", activity.ID)
	l.svcCtx.Redis.Setex(stockKey, fmt.Sprintf("%d", activity.TotalStock), expireSeconds)

	logx.Infof("synced activity %d to Redis: start=%d, stock=%d, ttl=%d seconds",
		activity.ID, activity.StartTime.UnixMilli(), activity.TotalStock, expireSeconds)
}
