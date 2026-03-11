package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/users"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除用户方法
func (l *DeleteUserLogic) DeleteUser(in *users.DeleteUserRequest) (*users.DeleteUserResponse, error) {
	// Soft delete user
	err := l.svcCtx.UsersModel.UpdateDeletebyId(l.ctx, int64(in.UserId), true)
	if err != nil {
		l.Logger.Errorw("delete user failed", logx.Field("err", err))
		return &users.DeleteUserResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	return &users.DeleteUserResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}, nil
}
