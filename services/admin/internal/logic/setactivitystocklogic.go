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

type SetActivityStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetActivityStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetActivityStockLogic {
	return &SetActivityStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetActivityStockLogic) SetActivityStock(in *adminpb.SetActivityStockRequest) (*adminpb.SetActivityStockResponse, error) {
	activity, err := db.GetActivityByID(l.svcCtx.DB, in.Id)
	if err != nil {
		return &adminpb.SetActivityStockResponse{
			StatusCode: 404,
			StatusMsg:  "activity not found",
		}, nil
	}

	activity.TotalStock = in.Stock

	if err := activity.Update(l.svcCtx.DB); err != nil {
		logx.Errorf("failed to update activity stock: %v", err)
		return &adminpb.SetActivityStockResponse{
			StatusCode: 500,
			StatusMsg:  "failed to update activity stock: " + err.Error(),
		}, nil
	}

	// Sync to Redis with TTL (活动结束后多保留1天)
	stockKey := fmt.Sprintf("act_%d_stock", activity.ID)
	expireSeconds := time.Until(activity.EndTime).Seconds()
	if expireSeconds < 0 {
		expireSeconds = biz.SeckillCacheTTL.Seconds()
	} else {
		expireSeconds += biz.SeckillCacheTTL.Seconds()
	}
	l.svcCtx.Redis.Setex(stockKey, fmt.Sprintf("%d", activity.TotalStock), int(expireSeconds))

	logx.Infof("updated activity %d stock in Redis: stock=%d, ttl=%d seconds",
		activity.ID, activity.TotalStock, int(expireSeconds))

	return &adminpb.SetActivityStockResponse{
		StatusCode: 200,
		StatusMsg:  "success",
	}, nil
}
