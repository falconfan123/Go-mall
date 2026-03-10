package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/userspb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个收货地址
func (l *GetAddressLogic) GetAddress(in *userspb.GetAddressRequest) (*userspb.GetAddressResponse, error) {
	// todo: add your logic here and delete this line

	return &userspb.GetAddressResponse{}, nil
}
