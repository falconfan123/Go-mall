package service

import (
	"context"
	"errors"
	"testing"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/order/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/order/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/order/internal/domain/repository"
	"github.com/stretchr/testify/require"
)

func TestOrderAppServiceCreateAndCancel(t *testing.T) {
	t.Parallel()

	repo := &stubOrderRepository{}
	service := NewOrderAppService(repo)

	createResp, err := service.CreateOrder(context.Background(), &dto.CreateOrderReq{
		PreOrderID:     "pre-1",
		UserID:         1,
		OriginalAmount: 1000,
		DiscountAmount: 100,
		PayableAmount:  900,
		Items: []*dto.OrderItemReq{
			{ProductID: 1001, ProductName: "product", ProductDesc: "desc", Quantity: 2, Price: 450},
		},
		Address: &dto.OrderAddressReq{
			AddressID:       10,
			RecipientName:   "fan",
			PhoneNumber:     "13800138000",
			Province:        "Zhejiang",
			City:            "Hangzhou",
			DetailedAddress: "No.1",
		},
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, createResp.StatusCode)
	require.NotNil(t, repo.savedOrder)

	repo.order = repo.savedOrder
	cancelResp, err := service.CancelOrder(context.Background(), &dto.CancelOrderReq{
		OrderID: repo.order.OrderID,
		Reason:  "changed mind",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, cancelResp.StatusCode)
	require.NotNil(t, repo.updatedOrder)
}

func TestOrderAppServicePayAndGet(t *testing.T) {
	t.Parallel()

	order := entity.NewOrder("ord-1", "pre-1", 1, "", 1000, 100, 900, 30)
	repo := &stubOrderRepository{order: order}
	service := NewOrderAppService(repo)

	payResp, err := service.PayOrder(context.Background(), &dto.PayOrderReq{
		OrderID:       "ord-1",
		PaymentMethod: 1,
		TransactionID: "tx-1",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, payResp.StatusCode)

	getResp, err := service.GetOrder(context.Background(), &dto.GetOrderReq{OrderID: "ord-1"})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, getResp.StatusCode)

	payInfoResp, err := service.GetOrder2Payment(context.Background(), &dto.GetOrder2PaymentReq{OrderID: "ord-1"})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, payInfoResp.StatusCode)
}

type stubOrderRepository struct {
	order        *entity.Order
	savedOrder   *entity.Order
	updatedOrder *entity.Order
}

func (s *stubOrderRepository) GetByID(context.Context, string) (*entity.Order, error) {
	if s.order == nil {
		return nil, errors.New("not found")
	}
	return s.order, nil
}

func (s *stubOrderRepository) GetByPreOrderID(context.Context, string) (*entity.Order, error) {
	return s.order, nil
}

func (s *stubOrderRepository) GetByUserID(context.Context, int64) ([]*entity.Order, error) {
	return []*entity.Order{s.order}, nil
}

func (s *stubOrderRepository) Save(_ context.Context, order *entity.Order) error {
	s.savedOrder = order
	return nil
}

func (s *stubOrderRepository) Update(_ context.Context, order *entity.Order) error {
	s.updatedOrder = order
	s.order = order
	return nil
}

func (s *stubOrderRepository) Delete(context.Context, string) error {
	return nil
}

func (s *stubOrderRepository) ListByUserID(context.Context, int64, *entity.OrderStatus, int, int) ([]*entity.Order, int64, error) {
	return []*entity.Order{s.order}, 1, nil
}

func (s *stubOrderRepository) ListByStatus(context.Context, entity.OrderStatus, int, int) ([]*entity.Order, int64, error) {
	return []*entity.Order{s.order}, 1, nil
}

func (s *stubOrderRepository) FindExpired(context.Context, int) ([]*entity.Order, error) {
	return nil, nil
}

var _ repository.OrderRepository = (*stubOrderRepository)(nil)
