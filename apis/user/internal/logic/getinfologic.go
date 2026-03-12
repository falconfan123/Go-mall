package logic

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/apis/user/internal/svc"
	"github.com/falconfan123/Go-mall/apis/user/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

// GetInfoLogic is the business logic for getinfo operations.
// GetInfoLogic is the business logic for GetInfoLogic operations.
type GetInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetInfoLogic creates a new instance.
// NewGetInfoLogic creates a new GetInfoLogic instance.
func NewGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoLogic {
	return &GetInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetInfo is a function.
//
//	does something.
func (l *GetInfoLogic) GetInfo(req *types.GetInfoRequest) (resp *types.GetInfoResponse, err error) {

	userID := l.ctx.Value(biz.UserIDKey).(uint32)

	getresp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &users.GetUserRequest{
		UserId: userID,
	})
	if err != nil {

		l.Logger.Errorw("call rpc getuser failed", logx.Field("err", err))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	} else if getresp.StatusMsg != "" {

		return nil, errors.New(int(getresp.StatusCode), getresp.StatusMsg)

	}
	resp = &types.GetInfoResponse{
		UserId:    int64(getresp.UserId),
		LogoutAt:  getresp.LogoutAt,
		CreatedAt: getresp.CreatedAt,
		UpdateAt:  getresp.UpdatedAt,
		Email:     getresp.Email,
		UserName:  getresp.UserName,
		Avatar:    getresp.AvatarUrl,
	}
	fmt.Println("resp:", resp)

	return resp, nil
}
