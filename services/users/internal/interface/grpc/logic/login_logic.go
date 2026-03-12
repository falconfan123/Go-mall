package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录方法
func (l *LoginLogic) Login(in *users.LoginRequest) (*users.LoginResponse, error) {
	// 调用应用服务处理登录逻辑
	req := &dto.LoginRequest{
		Email:    in.Email,
		Password: in.Password,
		IP:       in.Ip,
	}

	resp, err := l.svcCtx.AuthAppService.Login(l.ctx, req)
	if err != nil {
		l.Logger.Errorw("login failed", logx.Field("err", err))
		return &users.LoginResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	return &users.LoginResponse{
		StatusCode:   resp.StatusCode,
		StatusMsg:    resp.StatusMsg,
		UserId:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}
