package logic

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/services/activity/internal/svc"
	"github.com/falconfan123/Go-mall/services/activity/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckSeckillStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckSeckillStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckSeckillStatusLogic {
	return &CheckSeckillStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckSeckillStatus 检查用户购买状态
func (l *CheckSeckillStatusLogic) CheckSeckillStatus(in *pb.CheckStatusReq, userId int64) (*pb.CheckStatusResp, error) {
	activityId := in.ActivityId
	key := fmt.Sprintf("act_%d_bought", activityId)

	// 检查用户是否已购买
	isPurchased, err := l.svcCtx.Redis.Sismember(key, fmt.Sprintf("%d", userId))
	if err != nil {
		logx.Errorf("failed to check purchase status: %v", err)
		return &pb.CheckStatusResp{
			IsPurchased: false,
		}, nil
	}

	return &pb.CheckStatusResp{
		IsPurchased: isPurchased,
	}, nil
}