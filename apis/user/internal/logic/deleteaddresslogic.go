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

// DeleteAddressLogic is the business logic for deleteaddress operations.
// DeleteAddressLogic is the business logic for DeleteAddressLogic operations.
type DeleteAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteAddressLogic creates a new instance.
// NewDeleteAddressLogic creates a new DeleteAddressLogic instance.
func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteAddress is a function.
//
//	does something.
func (l *DeleteAddressLogic) DeleteAddress(req *types.DeleteAddressRequest) (resp *types.DeleteAddressResponse, err error) {
	userID := l.ctx.Value(biz.UserIDKey).(uint32)
	userIP := l.ctx.Value(biz.ClientIPKey).(string)
	DeleteAddResp, err := l.svcCtx.UserRPC.DeleteAddress(l.ctx, &users.DeleteAddressRequest{
		Ip:        userIP,
		UserId:    userID,
		AddressId: req.AddressID,
	})

	if err != nil {
		l.Logger.Errorw("调用 rpc 删除地址失败", logx.Field("err", err))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	} else if DeleteAddResp.StatusMsg != "" {

		return nil, errors.New(int(DeleteAddResp.StatusCode), DeleteAddResp.StatusMsg)

	}

	return
}
