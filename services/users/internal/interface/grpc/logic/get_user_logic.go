package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/dal/model/user"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/users"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户信息方法
func (l *GetUserLogic) GetUser(in *users.GetUserRequest) (*users.GetUserResponse, error) {
	u, err := l.svcCtx.UsersModel.FindOne(l.ctx, int64(in.UserId))
	if err != nil {
		if err == user.ErrNotFound {
			return &users.GetUserResponse{
				StatusCode: uint32(code.UserNotExistError),
				StatusMsg:  code.UserNotExistErrorMsg,
			}, nil
		}
		l.Logger.Errorw("get user failed", logx.Field("err", err))
		return &users.GetUserResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	return &users.GetUserResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserId:     uint32(u.UserId),
		Email:      u.Email.String,
		UserName:   u.Username.String,
		AvatarUrl:  u.AvatarUrl.String,
		CreatedAt:  u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  u.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
