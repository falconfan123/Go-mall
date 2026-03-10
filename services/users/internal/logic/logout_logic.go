package logic

import (
	"context"
	"time"

	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/userspb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登出方法
func (l *LogoutLogic) Logout(in *userspb.LogoutRequest) (*userspb.LogoutResponse, error) {
	// Invalidate token or update logout time in DB
	err := l.svcCtx.UsersModel.UpdateLogoutTime(l.ctx, int64(in.UserId), time.Now())
	if err != nil {
		l.Logger.Errorw("update logout time failed", logx.Field("err", err))
		// We might still return success if it's just a logging update
	}

	return &userspb.LogoutResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}, nil
}
