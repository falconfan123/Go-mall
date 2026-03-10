package logic

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/utils/token"
	"github.com/falconfan123/Go-mall/services/order/internal/svc"
	"github.com/falconfan123/Go-mall/services/order/order"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListOrders 分页查询订单列表
func (l *ListOrdersLogic) ListOrders(in *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	res := &order.ListOrdersResponse{}

	l.Logger.Infow("ListOrders called", logx.Field("in", in))
	fmt.Printf("ListOrders called. in=%+v\n", in)

	// Try to get user_id from metadata if missing
	if in.UserId == 0 {
		if md, ok := metadata.FromIncomingContext(l.ctx); ok {
			l.Logger.Infow("metadata received", logx.Field("md", md))
			fmt.Printf("metadata received: %+v\n", md)
			// Check injected user_id first
			userIds := md.Get("gateway-user-id")
			if len(userIds) == 0 {
				userIds = md.Get("user_id")
			}
			if len(userIds) > 0 {
				fmt.Printf("Found user_id in metadata: %v\n", userIds)
				if id, err := strconv.Atoi(userIds[0]); err == nil {
					in.UserId = uint32(id)
					fmt.Printf("Set in.UserId to %d\n", in.UserId)
				} else {
					fmt.Printf("Atoi error: %v\n", err)
				}
			}

			// Fallback to Authorization header
			if in.UserId == 0 {
				if auths := md.Get("authorization"); len(auths) > 0 {
					tokenStr := auths[0]
					if strings.HasPrefix(tokenStr, "Bearer ") {
						tokenStr = tokenStr[7:]
					}
					claims, err := token.ParseJWT(tokenStr)
					if err == nil {
						in.UserId = claims.UserID
					}
				}
			}
		}
	}

	// --------------- check ---------------
	if in.UserId == 0 {
		res.StatusCode = code.UserNotFound
		res.StatusMsg = code.UserNotFoundMsg
		return res, nil
	}

	if in.Pagination == nil {
		in.Pagination = &order.ListOrdersRequest_Pagination{
			Page:     1,
			PageSize: 10,
		}
	}

	if in.Pagination.PageSize <= 0 || in.Pagination.PageSize > biz.MaxPageSize {
		in.Pagination.PageSize = biz.MaxPageSize
	}
	if in.Pagination.Page <= 0 {
		in.Pagination.Page = 1
	}
	orderList, err := l.svcCtx.OrderModel.GetOrdersByUserID(l.ctx, int32(in.UserId), in.Pagination.Page, in.Pagination.PageSize)
	if err != nil {
		l.Logger.Errorw("call svcCtx.OrderModel.GetOrdersByUserID failed", logx.Field("err", err))
		res.StatusCode = code.ServerError
		res.StatusMsg = code.ServerErrorMsg
		return res, nil
	}
	res.Orders = make([]*order.Order, len(orderList))
	for i, o := range orderList {
		orderResp := convertToOrderResp(o)
		// 查询订单关联的订单项
		items, err := l.svcCtx.OrderItemModel.QueryOrderItemsByOrderID(l.ctx, o.OrderId)
		if err != nil {
			l.Logger.Errorw("call svcCtx.OrderItemModel.QueryOrderItemsByOrderID failed", logx.Field("err", err))
			continue
		}
		orderResp.Items = convertToOrderItemResp(items)
		res.Orders[i] = orderResp
	}

	return res, nil
}
