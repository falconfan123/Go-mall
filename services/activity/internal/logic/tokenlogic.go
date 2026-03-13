package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/services/activity/internal/svc"
	"github.com/falconfan123/Go-mall/services/activity/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type TokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TokenLogic {
	return &TokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TokenLogic) Token(in *pb.TokenReq, userId int64) (*pb.TokenResp, error) {
	activityId := in.ActivityId
	salt := "seckill_salt_2026"

	// 生成 path_key: md5(userId + activityId + salt)
	raw := fmt.Sprintf("%d_%d_%s", userId, activityId, salt)
	hash := md5.Sum([]byte(raw))
	pathKey := hex.EncodeToString(hash[:])

	// 计算活动开始时间（这里应该从 Redis 获取实际的活动开始时间）
	// 假设活动开始时间已存储在 Redis 中
	// 提前 N 秒生成 token
	now := time.Now().UnixMilli()

	// 从 Redis 获取活动开始时间
	startTimeKey := fmt.Sprintf("act_%d_start", activityId)
	startTime, err := l.svcCtx.Redis.Get(startTimeKey)
	if err != nil || startTime == "" {
		// 如果没有设置活动开始时间，返回错误
		logx.Errorf("activity %d not found or not started", activityId)
		return nil, fmt.Errorf("activity not found or not started")
	}

	// 检查是否在活动开始前 N 秒内
	startTimeMs := 0
	fmt.Sscanf(startTime, "%d", &startTimeMs)
	advanceMs := int64(l.svcCtx.Config.Activity.AdvanceSeconds * 1000)

	if now < int64(startTimeMs)-advanceMs {
		logx.Errorf("activity %d not yet available, now: %d, start: %d", activityId, now, startTimeMs)
		return nil, fmt.Errorf("activity not yet available")
	}

	// 存储 path_key 到 Redis，有效期 1 分钟
	key := fmt.Sprintf("act_%d_path_%d", activityId, userId)
	expire := l.svcCtx.Config.Activity.TokenExpire
	err = l.svcCtx.Redis.Setex(key, pathKey, expire)
	if err != nil {
		logx.Errorf("failed to set path_key: %v", err)
		return nil, err
	}

	return &pb.TokenResp{
		PathKey:   pathKey,
		ExpiresAt: now + int64(expire*1000),
	}, nil
}
