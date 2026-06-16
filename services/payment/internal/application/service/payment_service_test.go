package service

import (
	"context"
	"errors"
	"testing"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/payment/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/payment/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/payment/internal/domain/repository"
	"github.com/stretchr/testify/require"
)

func TestPaymentAppServiceCreateAndGet(t *testing.T) {
	t.Parallel()

	repo := &stubPaymentRepository{}
	service := NewPaymentAppService(repo)

	createResp, err := service.CreatePayment(context.Background(), &dto.CreatePaymentReq{
		UserID:        1,
		OrderID:       "ord-1",
		PaymentMethod: 3,
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, createResp.StatusCode)
	require.NotNil(t, repo.savedPayment)

	repo.payment = repo.savedPayment
	getResp, err := service.GetPayment(context.Background(), &dto.GetPaymentReq{PaymentID: repo.payment.PaymentID})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, getResp.StatusCode)
}

func TestPaymentAppServiceIdempotencyAndUnsupportedMethod(t *testing.T) {
	t.Parallel()

	existing := entity.NewPayment("pay-1", "pre-1", "ord-1", 1, 1000, 900, entity.PaymentMethodStripe, "url", 30)
	service := NewPaymentAppService(&stubPaymentRepository{paymentByOrder: existing})

	resp, err := service.CreatePayment(context.Background(), &dto.CreatePaymentReq{
		UserID:        1,
		OrderID:       "ord-1",
		PaymentMethod: 3,
	})
	require.NoError(t, err)
	require.EqualValues(t, code.PaymentExist, resp.StatusCode)

	service = NewPaymentAppService(&stubPaymentRepository{})
	resp, err = service.CreatePayment(context.Background(), &dto.CreatePaymentReq{
		UserID:        1,
		OrderID:       "ord-2",
		PaymentMethod: 99,
	})
	require.NoError(t, err)
	require.EqualValues(t, code.PaymentMethodNotSupport, resp.StatusCode)
}

type stubPaymentRepository struct {
	payment        *entity.Payment
	paymentByOrder *entity.Payment
	savedPayment   *entity.Payment
}

func (s *stubPaymentRepository) GetByID(context.Context, string) (*entity.Payment, error) {
	if s.payment == nil {
		return nil, entity.ErrPaymentNotFound
	}
	return s.payment, nil
}

func (s *stubPaymentRepository) GetByOrderID(context.Context, string) (*entity.Payment, error) {
	if s.paymentByOrder == nil {
		return nil, entity.ErrPaymentNotFound
	}
	return s.paymentByOrder, nil
}

func (s *stubPaymentRepository) GetByPreOrderID(context.Context, string) (*entity.Payment, error) {
	if s.payment == nil {
		return nil, errors.New("not found")
	}
	return s.payment, nil
}

func (s *stubPaymentRepository) Save(_ context.Context, payment *entity.Payment) error {
	s.savedPayment = payment
	return nil
}

func (s *stubPaymentRepository) Update(context.Context, *entity.Payment) error {
	return nil
}

func (s *stubPaymentRepository) Delete(context.Context, string) error {
	return nil
}

func (s *stubPaymentRepository) ListByUserID(context.Context, int64, *entity.PaymentStatus, int, int) ([]*entity.Payment, int64, error) {
	if s.payment == nil {
		return nil, 0, nil
	}
	return []*entity.Payment{s.payment}, 1, nil
}

func (s *stubPaymentRepository) ListByStatus(context.Context, entity.PaymentStatus, int, int) ([]*entity.Payment, int64, error) {
	return nil, 0, nil
}

func (s *stubPaymentRepository) FindExpired(context.Context, int) ([]*entity.Payment, error) {
	return nil, nil
}

var _ repository.PaymentRepository = (*stubPaymentRepository)(nil)
