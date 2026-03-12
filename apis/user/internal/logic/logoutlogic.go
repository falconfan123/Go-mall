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

// LogoutLogic is the business logic for logout operations.
// LogoutLogic is the business logic for LogoutLogic operations.
type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLogoutLogic creates a new instance.
// NewLogoutLogic creates a new LogoutLogic instance.
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Logout is a function.
//
//	does something.
func (l *LogoutLogic) Logout(req *types.LogoutRequest) (resp *types.LogoutResponse, err error) {

	userID := l.ctx.Value(biz.UserIDKey).(uint32)

	logoutrep, err := l.svcCtx.UserRPC.Logout(l.ctx, &users.LogoutRequest{

		UserId: userID,
	})
	if err != nil {

		l.Logger.Errorw("call rpc logout failed", logx.Field("err", err))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	} else if logoutrep.StatusMsg != "" {

		return nil, errors.New(int(logoutrep.StatusCode), logoutrep.StatusMsg)

	}

	resp = &types.LogoutResponse{
		Logout_at: logoutrep.LogoutTime,
	}

	return
}
