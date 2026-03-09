package order

import (
	"context"

	"jijizhazha1024/go-mall/apis/order/internal/svc"
	"jijizhazha1024/go-mall/apis/order/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderDetailLogic) OrderDetail(req *types.GetOrderReq) (resp *types.OrderDetailResp, err error) {
	// todo: add your logic here and delete this line

	return
}
