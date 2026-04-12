package logic

import (
	"context"
	"time"

	"github.com/falconfan123/Go-mall/services/system/internal/svc"
	system "github.com/falconfan123/Go-mall/services/system/pb"
)

type TimeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TimeLogic {
	return &TimeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TimeLogic) Time(in *system.TimeReq) (*system.TimeResp, error) {
	return &system.TimeResp{
		Now: time.Now().UnixMilli(),
	}, nil
}
