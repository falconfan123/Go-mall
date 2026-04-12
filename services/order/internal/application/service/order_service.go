package service

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/order/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/order/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/order/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/order/internal/domain/repository"
)

// OrderAppService 订单应用服务
type OrderAppService struct {
	orderRepo repository.OrderRepository
}

// NewOrderAppService 创建订单应用服务
func NewOrderAppService(orderRepo repository.OrderRepository) *OrderAppService {
	return &OrderAppService{
		orderRepo: orderRepo,
	}
}

// CreateOrder 创建订单
func (s *OrderAppService) CreateOrder(ctx context.Context, req *dto.CreateOrderReq) (*dto.CreateOrderResp, error) {
	// 1. 生成订单ID
	orderID := s.generateOrderID(req.UserID)

	// 2. 创建订单聚合根
	orderAgg := aggregate.NewOrderAggregate(
		orderID,
		req.PreOrderID,
		req.UserID,
		req.CouponID,
		req.OriginalAmount,
		req.DiscountAmount,
		req.PayableAmount,
		30, // 30分钟过期
	)

	// 3. 添加订单项
	for _, item := range req.Items {
		if err := orderAgg.AddItem(
			item.ProductID,
			item.Quantity,
			item.Price,
			item.ProductName,
			item.ProductDesc,
		); err != nil {
			return &dto.CreateOrderResp{
				StatusCode: code.Fail,
				StatusMsg:  err.Error(),
			}, err
		}
	}

	// 4. 设置地址快照
	if req.Address != nil {
		orderAgg.SetAddress(
			req.Address.AddressID,
			req.Address.RecipientName,
			req.Address.PhoneNumber,
			req.Address.Province,
			req.Address.City,
			req.Address.DetailedAddress,
		)
	}

	// 5. 保存订单
	order := orderAgg.GetOrder()
	if err := s.orderRepo.Save(ctx, order); err != nil {
		return &dto.CreateOrderResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.CreateOrderResp{
		OrderID:    orderID,
		StatusCode: code.Success,
		StatusMsg:  "order created successfully",
	}, nil
}

// GetOrder 获取订单详情
func (s *OrderAppService) GetOrder(ctx context.Context, req *dto.GetOrderReq) (*dto.GetOrderResp, error) {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return &dto.GetOrderResp{
			StatusCode: code.OrderNotExist,
			StatusMsg:  code.OrderNotExistMsg,
		}, nil
	}

	return &dto.GetOrderResp{
		Order:      s.convertOrderToDTO(order),
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// ListOrders 查询订单列表
func (s *OrderAppService) ListOrders(ctx context.Context, req *dto.ListOrdersReq) (*dto.ListOrdersResp, error) {
	var status *entity.OrderStatus
	if req.Status != nil {
		st := entity.OrderStatus(*req.Status)
		status = &st
	}

	orders, total, err := s.orderRepo.ListByUserID(ctx, req.UserID, status, req.Page, req.PageSize)
	if err != nil {
		return &dto.ListOrdersResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	items := make([]*dto.OrderListItemDTO, 0, len(orders))
	for _, o := range orders {
		items = append(items, &dto.OrderListItemDTO{
			OrderID:       o.OrderID,
			TotalAmount:   o.PayableAmount,
			OrderStatus:   int64(o.OrderStatus),
			PaymentStatus: int64(o.PaymentStatus),
			ItemCount:     len(o.Items),
			CreatedAt:     o.CreatedAt.Unix(),
		})
	}

	return &dto.ListOrdersResp{
		Orders:     items,
		TotalCount: total,
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// CancelOrder 取消订单
func (s *OrderAppService) CancelOrder(ctx context.Context, req *dto.CancelOrderReq) (*dto.CancelOrderResp, error) {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return &dto.CancelOrderResp{
			StatusCode: code.OrderNotExist,
			StatusMsg:  code.OrderNotExistMsg,
		}, nil
	}

	orderAgg := aggregate.LoadOrder(order)
	if err := orderAgg.Cancel(req.Reason); err != nil {
		return &dto.CancelOrderResp{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, nil
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return &dto.CancelOrderResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.CancelOrderResp{
		StatusCode: code.Success,
		StatusMsg:  "order canceled successfully",
	}, nil
}

// PayOrder 支付订单
func (s *OrderAppService) PayOrder(ctx context.Context, req *dto.PayOrderReq) (*dto.PayOrderResp, error) {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return &dto.PayOrderResp{
			StatusCode: code.OrderNotExist,
			StatusMsg:  code.OrderNotExistMsg,
		}, nil
	}

	orderAgg := aggregate.LoadOrder(order)
	if err := orderAgg.Pay(req.PaymentMethod, req.TransactionID); err != nil {
		return &dto.PayOrderResp{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, nil
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return &dto.PayOrderResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.PayOrderResp{
		StatusCode: code.Success,
		StatusMsg:  "order paid successfully",
	}, nil
}

// GetOrder2Payment 获取订单支付信息
func (s *OrderAppService) GetOrder2Payment(ctx context.Context, req *dto.GetOrder2PaymentReq) (*dto.GetOrder2PaymentResp, error) {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return &dto.GetOrder2PaymentResp{
			StatusCode: code.OrderNotExist,
			StatusMsg:  code.OrderNotExistMsg,
		}, nil
	}

	return &dto.GetOrder2PaymentResp{
		OrderID:       order.OrderID,
		PayableAmount: order.PayableAmount,
		StatusCode:    code.Success,
		StatusMsg:     "success",
	}, nil
}

// 辅助方法
func (s *OrderAppService) generateOrderID(userID int64) string {
	return fmt.Sprintf("ORD%d%d", userID, time.Now().Unix())
}

func (s *OrderAppService) convertOrderToDTO(order *entity.Order) *dto.OrderDTO {
	items := make([]*dto.OrderItemDTO, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &dto.OrderItemDTO{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			ProductDesc: item.ProductDesc,
			Quantity:    item.Quantity,
			Price:       item.Price,
			TotalPrice:  item.TotalPrice(),
		})
	}

	var address *dto.OrderAddressDTO
	if order.Address != nil {
		address = &dto.OrderAddressDTO{
			AddressID:       order.Address.AddressID,
			RecipientName:   order.Address.RecipientName,
			PhoneNumber:     order.Address.PhoneNumber,
			Province:        order.Address.Province,
			City:            order.Address.City,
			DetailedAddress: order.Address.DetailedAddress,
		}
	}

	var paidAt *int64
	if order.PaidAt != nil {
		pt := order.PaidAt.Unix()
		paidAt = &pt
	}

	return &dto.OrderDTO{
		OrderID:        order.OrderID,
		PreOrderID:     order.PreOrderID,
		UserID:         order.UserID,
		CouponID:       order.CouponID,
		OriginalAmount: order.OriginalAmount,
		DiscountAmount: order.DiscountAmount,
		PayableAmount:  order.PayableAmount,
		PaidAmount:     order.PaidAmount,
		OrderStatus:    int64(order.OrderStatus),
		PaymentStatus:  int64(order.PaymentStatus),
		PaymentMethod:  order.PaymentMethod,
		TransactionID:  order.TransactionID,
		ExpireTime:     order.ExpireTime.Unix(),
		Items:          items,
		Address:        address,
		CreatedAt:      order.CreatedAt.Unix(),
		PaidAt:         paidAt,
	}
}
