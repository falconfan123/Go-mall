package logic

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	users "github.com/falconfan123/Go-mall/services/users/pb"

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
	username := in.Username
	if username == "" {
		username = in.Email // 默认用户名为邮箱
	}
	req := &dto.RegisterRequest{
		Email:           in.Email,
		Username:        username,
		Password:        in.Password,
		ConfirmPassword: in.ConfirmPassword,
		IP:              in.Ip,
		DeviceID:        in.DeviceId,
	}

	resp, err := l.svcCtx.AuthAppService.Register(l.ctx, req)
	if err != nil {
		l.Logger.Errorw("register failed",
			logx.Field("err", err),
			logx.Field("username", req.Username),
			logx.Field("email", req.Email),
		)
		statusMsg := code.ServerErrorMsg
		if resp != nil && resp.StatusMsg != "" {
			statusMsg = fmt.Sprintf("%s: %s", resp.StatusMsg, err.Error())
		} else {
			statusMsg = fmt.Sprintf("%s: %v", code.ServerErrorMsg, err)
		}
		return &users.RegisterResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  statusMsg,
		}, nil
	}

	return &users.RegisterResponse{
		StatusCode:     resp.StatusCode,
		StatusMsg:      resp.StatusMsg,
		UserId:         resp.UserID,
		ShortToken:     resp.ShortToken,
		LongToken:      resp.LongToken,
		ShortExpiresIn: resp.ShortExpiresIn,
		LongExpiresIn:  resp.LongExpiresIn,
	}, nil
}
