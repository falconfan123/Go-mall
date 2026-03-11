package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/users"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册方法
func (l *RegisterLogic) Register(in *users.RegisterRequest) (*users.RegisterResponse, error) {
	// 调用应用服务处理注册逻辑
	req := &dto.RegisterRequest{
		Email:           in.Email,
		Password:        in.Password,
		ConfirmPassword: in.ConfirmPassword,
		Username:        in.Email, // 默认用户名为邮箱
		IP:              in.Ip,
	}

	resp, err := l.svcCtx.AuthAppService.Register(l.ctx, req)
	if err != nil {
		l.Logger.Errorw("register failed", logx.Field("err", err))
		return &users.RegisterResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	return &users.RegisterResponse{
		StatusCode:   resp.StatusCode,
		StatusMsg:    resp.StatusMsg,
		UserId:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}
