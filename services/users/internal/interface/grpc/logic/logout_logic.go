package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/pb"

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
func (l *LogoutLogic) Logout(in *users.LogoutRequest) (*users.LogoutResponse, error) {
	// 调用应用服务处理登出逻辑
	req := &dto.LogoutRequest{
		UserID: in.UserId,
		IP:     "", // 从ctx获取IP，这里暂时留空
	}

	resp, err := l.svcCtx.AuthAppService.Logout(l.ctx, req)
	if err != nil {
		l.Logger.Errorw("logout failed", logx.Field("err", err))
		return &users.LogoutResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	return &users.LogoutResponse{
		StatusCode: resp.StatusCode,
		StatusMsg:  resp.StatusMsg,
	}, nil
}
