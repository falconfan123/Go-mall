package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/checkout/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/repository"
	"github.com/stretchr/testify/require"
)

func TestCheckoutAppServicePrepareAndGet(t *testing.T) {
	t.Parallel()

	repo := &stubCheckoutRepository{}
	service := NewCheckoutAppService(repo)

	prepareResp, err := service.PrepareCheckout(context.Background(), &dto.PrepareCheckoutReq{
		UserID:    1,
		AddressID: 10,
		Items: []*dto.CheckoutItemReq{
			{ProductID: 1001, ProductName: "product", ProductImage: "img", Quantity: 2, Price: 300},
		},
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, prepareResp.StatusCode)
	require.NotNil(t, repo.savedCheckout)

	repo.checkout = repo.savedCheckout
	getResp, err := service.GetCheckoutDetail(context.Background(), &dto.GetCheckoutDetailReq{PreOrderID: repo.checkout.PreOrderID})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, getResp.StatusCode)
}

func TestCheckoutAppServiceReleaseAndList(t *testing.T) {
	t.Parallel()

	checkout := entity.NewCheckout("pre-1", 1, 10, 1000, 30)
	checkout.ExpireTime = time.Now().Add(-time.Minute)
	repo := &stubCheckoutRepository{checkout: checkout}
	service := NewCheckoutAppService(repo)

	releaseResp, err := service.ReleaseCheckout(context.Background(), &dto.ReleaseCheckoutReq{PreOrderID: "pre-1"})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, releaseResp.StatusCode)

	listResp, err := service.ListCheckouts(context.Background(), &dto.ListCheckoutReq{
		UserID:   1,
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, listResp.StatusCode)
}

type stubCheckoutRepository struct {
	checkout      *entity.Checkout
	savedCheckout *entity.Checkout
}

func (s *stubCheckoutRepository) GetByID(context.Context, string) (*entity.Checkout, error) {
	if s.checkout == nil {
		return nil, errors.New("not found")
	}
	return s.checkout, nil
}

func (s *stubCheckoutRepository) GetByUserID(context.Context, int64) ([]*entity.Checkout, error) {
	return []*entity.Checkout{s.checkout}, nil
}

func (s *stubCheckoutRepository) Save(_ context.Context, checkout *entity.Checkout) error {
	s.savedCheckout = checkout
	return nil
}

func (s *stubCheckoutRepository) Update(_ context.Context, checkout *entity.Checkout) error {
	s.checkout = checkout
	return nil
}

func (s *stubCheckoutRepository) Delete(context.Context, string) error {
	return nil
}

func (s *stubCheckoutRepository) ListByUserID(context.Context, int64, int, int) ([]*entity.Checkout, int64, error) {
	if s.checkout == nil {
		return nil, 0, nil
	}
	return []*entity.Checkout{s.checkout}, 1, nil
}

func (s *stubCheckoutRepository) FindExpired(context.Context, int) ([]*entity.Checkout, error) {
	return nil, nil
}

func (s *stubCheckoutRepository) DecreaseStock(context.Context, []*entity.CheckoutItem) error {
	return nil
}

func (s *stubCheckoutRepository) IncreaseStock(context.Context, []*entity.CheckoutItem) error {
	return nil
}

var _ repository.CheckoutRepository = (*stubCheckoutRepository)(nil)
