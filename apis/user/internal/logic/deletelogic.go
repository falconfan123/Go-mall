package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/apis/user/internal/svc"
	"github.com/falconfan123/Go-mall/apis/user/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

// DeleteLogic is the business logic for delete operations.
// DeleteLogic is the business logic for DeleteLogic operations.
type DeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteLogic creates a new instance.
// NewDeleteLogic creates a new DeleteLogic instance.
func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Delete is a function.
//
//	does something.
func (l *DeleteLogic) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {

	userID := l.ctx.Value(biz.UserIDKey).(uint32)
	userIP := l.ctx.Value(biz.ClientIPKey).(string)

	deleteresp, err := l.svcCtx.UserRPC.DeleteUser(l.ctx, &users.DeleteUserRequest{

		UserId: uint32(userID),
		Ip:     userIP,
	})
	if err != nil {

		l.Logger.Errorw("call rpc deleteuser failed", logx.Field("err", err))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	} else if deleteresp.StatusMsg != "" {

		return nil, errors.New(int(deleteresp.StatusCode), deleteresp.StatusMsg)

	}
	resp = &types.DeleteResponse{}

	return
}
