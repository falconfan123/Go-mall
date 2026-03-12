package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/order/internal/svc"
	"github.com/falconfan123/Go-mall/services/order/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderRequest) (*order.OrderDetailResponse, error) {
	// todo: add your logic here and delete this line

	return &order.OrderDetailResponse{}, nil
}
