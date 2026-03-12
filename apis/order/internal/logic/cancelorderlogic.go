package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/order/pb"
	xerrors "github.com/zeromicro/x/errors"

	"github.com/falconfan123/Go-mall/apis/order/internal/svc"
	"github.com/falconfan123/Go-mall/apis/order/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// CancelOrderLogic is the business logic for cancelorder operations.
// CancelOrderLogic is the business logic for CancelOrderLogic operations.
type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCancelOrderLogic creates a new instance.
// NewCancelOrderLogic creates a new CancelOrderLogic instance.
func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CancelOrder is a function.
//
//	does something.
func (l *CancelOrderLogic) CancelOrder(req *types.CancelOrderReq) (resp *types.CancelOrderResp, err error) {
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, xerrors.New(code.AuthBlank, code.AuthBlankMsg)
	}
	res, err := l.svcCtx.OrderRPC.CancelOrder(l.ctx, &order.CancelOrderRequest{
		OrderId:      req.OrderID,
		UserId:       userID,
		CancelReason: req.CancelReason,
		Initiative:   true,
	})
	if err != nil {
		l.Logger.Errorw("call rpc GetOrder failed", logx.Field("err", err))
		return nil, xerrors.New(code.ServerError, code.ServerErrorMsg)
	}
	if res.StatusCode != code.Success {
		return nil, xerrors.New(int(res.StatusCode), res.StatusMsg)
	}
	resp = &types.CancelOrderResp{
		OrderID: req.OrderID,
	}
	return
}
