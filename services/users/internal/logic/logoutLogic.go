package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/users"

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
	// todo: add your logic here and delete this line

	return &userspb.LogoutResponse{}, nil
}
