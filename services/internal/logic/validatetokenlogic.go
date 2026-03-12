package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/auths/auths/auths"
	"github.com/falconfan123/Go-mall/services/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ValidateToken 验证长短令牌（用于网关）
func (l *ValidateTokenLogic) ValidateToken(in *auths.AuthValidateReq) (*auths.AuthValidateRes, error) {
	// todo: add your logic here and delete this line

	return &auths.AuthValidateRes{}, nil
}
