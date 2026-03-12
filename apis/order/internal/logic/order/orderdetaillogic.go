package order

import (
	"context"

	"github.com/falconfan123/Go-mall/apis/order/internal/svc"
	"github.com/falconfan123/Go-mall/apis/order/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// OrderDetailLogic is the business logic for OrderDetailLogic operations.
type OrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOrderDetailLogic creates a new OrderDetailLogic instance.
func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// does something.
func (l *OrderDetailLogic) OrderDetail(req *types.GetOrderReq) (resp *types.OrderDetailResp, err error) {
	// todo: add your logic here and delete this line

	return
}
