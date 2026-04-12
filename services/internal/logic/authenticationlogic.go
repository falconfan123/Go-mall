package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/auths/auths/auths"
	"github.com/falconfan123/Go-mall/services/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthenticationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthenticationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthenticationLogic {
	return &AuthenticationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Authentication 验证用户token合法
func (l *AuthenticationLogic) Authentication(in *auths.AuthReq) (*auths.AuthsRes, error) {
	// todo: add your logic here and delete this line

	return &auths.AuthsRes{}, nil
}
