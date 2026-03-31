package logic

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	admin "github.com/falconfan123/Go-mall/services/admin/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearSeckillRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearSeckillRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearSeckillRecordLogic {
	return &ClearSeckillRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ClearSeckillRecord 清除秒杀购买记录
func (l *ClearSeckillRecordLogic) ClearSeckillRecord(in *admin.ClearSeckillRecordRequest) (*admin.ClearSeckillRecordResponse, error) {
	activityId := in.ActivityId

	if activityId <= 0 {
		return &admin.ClearSeckillRecordResponse{
			StatusCode: 1,
			StatusMsg:  "无效的活动ID",
		}, nil
	}

	var clearedCount int64

	// 清除购买记录
	if in.UserId > 0 {
		// 清除指定用户的购买记录
		key := fmt.Sprintf("act_%d_bought", activityId)
		removed, err := l.svcCtx.Redis.SremCtx(l.ctx, key, fmt.Sprintf("%d", in.UserId))
		if err != nil {
			logx.Errorf("failed to clear seckill record for user %d: %v", in.UserId, err)
			return &admin.ClearSeckillRecordResponse{
				StatusCode: 1,
				StatusMsg:  "清除记录失败",
			}, nil
		}
		clearedCount = int64(removed)
		logx.Infof("cleared seckill record for user %d, activity %d", in.UserId, activityId)
	} else {
		// 清除所有用户的购买记录
		key := fmt.Sprintf("act_%d_bought", activityId)
		exists, _ := l.svcCtx.Redis.ExistsCtx(l.ctx, key)
		if exists {
			// 获取 Set 中所有成员的数量
			count, err := l.svcCtx.Redis.ScardCtx(l.ctx, key)
			if err == nil {
				clearedCount = int64(count)
			}
			// 删除整个 Set
			_, err = l.svcCtx.Redis.DelCtx(l.ctx, key)
			if err != nil {
				logx.Errorf("failed to clear all seckill records for activity %d: %v", activityId, err)
				return &admin.ClearSeckillRecordResponse{
					StatusCode: 1,
					StatusMsg:  "清除记录失败",
				}, nil
			}
		}
		logx.Infof("cleared all seckill records for activity %d, count: %d", activityId, clearedCount)
	}

	// 如果需要重置库存
	if in.ClearStock {
		stockKey := fmt.Sprintf("act_%d_stock", activityId)
		// 先删除旧的库存 key
		l.svcCtx.Redis.DelCtx(l.ctx, stockKey)
		logx.Infof("reset stock for activity %d", activityId)
	}

	return &admin.ClearSeckillRecordResponse{
		StatusCode:   0,
		StatusMsg:    "success",
		ClearedCount: clearedCount,
	}, nil
}
