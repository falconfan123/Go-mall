package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户方法
func (l *UpdateUserLogic) UpdateUser(in *users.UpdateUserRequest) (*users.UpdateUserResponse, error) {
	// Update user info
	// Use UsrName instead of Username as per proto definition
	err := l.svcCtx.UsersModel.UpdateUserNameandUrl(l.ctx, int64(in.UserId), in.UsrName, in.AvatarUrl)
	if err != nil {
		l.Logger.Errorw("update user failed", logx.Field("err", err))
		return &users.UpdateUserResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	return &users.UpdateUserResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}, nil
}
