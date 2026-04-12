package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/auths/auths/auths"
	"github.com/falconfan123/Go-mall/services/internal/svc"

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

// Logout 登出（删除Session）
func (l *LogoutLogic) Logout(in *auths.LogoutReq) (*auths.LogoutRes, error) {
	// todo: add your logic here and delete this line

	return &auths.LogoutRes{}, nil
}
