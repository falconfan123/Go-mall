package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/order/internal/closeorder"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	order "github.com/falconfan123/Go-mall/services/order/pb"

	"github.com/falconfan123/Go-mall/services/order/internal/svc"
)

type CloseExpiredOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseExpiredOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseExpiredOrderLogic {
	return &CloseExpiredOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CloseExpiredOrder 关闭超时未支付订单（对账扫描器 scan2 与 delay consumer 共用业务，
// design 决策 7；业务本体在 internal/closeorder 包，避免 delay→logic→svc→delay 导入环）。
//
// 幂等响应语义：订单非 Created（已支付/已关闭）→ Success + "already closed, skip"
// （重试安全）；业务失败 → 返回错误（调用方重试，连续失败由调用方产出 exception）。
func (l *CloseExpiredOrderLogic) CloseExpiredOrder(in *order.CloseExpiredOrderRequest) (*order.EmptyRes, error) {
	biz := closeorder.New(l.svcCtx.OrderModel, l.svcCtx.OrderItemModel, l.svcCtx.InventoryRpc,
		l.svcCtx.Model, l.Logger)
	statusCode, statusMsg, err := biz.Close(l.ctx, in.OrderId, in.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "关闭超时订单失败")
	}
	return &order.EmptyRes{StatusCode: statusCode, StatusMsg: statusMsg}, nil
}
