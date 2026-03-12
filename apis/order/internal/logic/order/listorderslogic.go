package order

import (
	"context"

	"strconv"

	"github.com/falconfan123/Go-mall/apis/order/internal/svc"
	"github.com/falconfan123/Go-mall/apis/order/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/order/order"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type ListOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrdersLogic) ListOrders(req *types.ListOrdersReq) (resp *types.ListOrdersResp, err error) {
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, errors.New(code.AuthBlank, code.AuthBlankMsg)
	}

	rpcReq := &order.ListOrdersRequest{
		UserId: userID,
		Pagination: &order.ListOrdersRequest_Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}
	if len(req.Statuses) > 0 {
		statusFilter := &order.ListOrdersRequest_OrderStatusFilter{
			Statuses: make([]order.OrderStatus, len(req.Statuses)),
		}
		for i, s := range req.Statuses {
			statusFilter.Statuses[i] = order.OrderStatus(s)
		}
		rpcReq.StatusFilter = statusFilter
	}

	rpcResp, err := l.svcCtx.OrderRpc.ListOrders(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("ListOrders rpc failed: %v", err)
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	}
	if rpcResp.StatusCode != code.Success {
		return nil, errors.New(int(rpcResp.StatusCode), rpcResp.StatusMsg)
	}

	resp = &types.ListOrdersResp{
		Orders: make([]types.OrderResp, len(rpcResp.Orders)),
	}

	for i, o := range rpcResp.Orders {
		resp.Orders[i] = types.OrderResp{
			OrderID:        o.OrderId,
			PreOrderID:     o.PreOrderId,
			UserID:         o.UserId,
			PaymentMethod:  int32(o.PaymentMethod),
			TransactionID:  o.TransactionId,
			PaidAt:         o.PaidAt,
			OriginalAmount: strconv.FormatInt(o.OriginalAmount, 10),
			DiscountAmount: strconv.FormatInt(o.DiscountAmount, 10),
			PayableAmount:  strconv.FormatInt(o.PayableAmount, 10),
			PaidAmount:     strconv.FormatInt(o.PaidAmount, 10),
			OrderStatus:    int32(o.OrderStatus),
			PaymentStatus:  int32(o.PaymentStatus),
			Reason:         o.Reason,
			ExpireTime:     o.ExpireTime,
			CreatedAt:      o.CreatedAt,
			UpdatedAt:      o.UpdatedAt,
			Items:          make([]types.OrderItemResp, len(o.Items)),
		}
		for j, item := range o.Items {
			resp.Orders[i].Items[j] = types.OrderItemResp{
				ItemID:      item.ItemId,
				ProductID:   item.ProductId,
				Quantity:    item.Quantity,
				ProductName: item.ProductName,
				ProductDesc: item.ProductDesc,
				UnitPrice:   strconv.FormatInt(item.UnitPrice, 10),
			}
		}
	}

	return resp, nil
}
