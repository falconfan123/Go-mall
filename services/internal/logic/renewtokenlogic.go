package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/auths/auths/auths"
	"github.com/falconfan123/Go-mall/services/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RenewTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRenewTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RenewTokenLogic {
	return &RenewTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RenewToken 续期身份
func (l *RenewTokenLogic) RenewToken(in *auths.AuthRenewalReq) (*auths.AuthRenewalRes, error) {
	// todo: add your logic here and delete this line

	return &auths.AuthRenewalRes{}, nil
}
