package service

import (
	"context"
	"testing"

	"github.com/falconfan123/Go-mall/services/carts/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/repository"
	"github.com/stretchr/testify/require"
)

func TestCartAppServiceAddUpdateAndGet(t *testing.T) {
	t.Parallel()

	repo := &stubCartRepository{}
	service := NewCartAppService(repo)

	err := service.AddItem(context.Background(), &dto.AddItemReq{
		UserID:       1,
		ProductID:    1001,
		ProductName:  "product",
		ProductImage: "img",
		ProductPrice: 99.9,
		Quantity:     2,
	})
	require.NoError(t, err)
	require.NotNil(t, repo.savedCart)

	repo.cart = repo.savedCart
	err = service.UpdateItemQuantity(context.Background(), &dto.UpdateItemQuantityReq{
		UserID:    1,
		ProductID: 1001,
		Quantity:  3,
	})
	require.NoError(t, err)

	cart, err := service.GetCart(context.Background(), &dto.GetCartReq{UserID: 1})
	require.NoError(t, err)
	require.Equal(t, int32(3), cart.TotalQuantity)
	require.Len(t, cart.Items, 1)
}

func TestCartAppServiceCheckAndClear(t *testing.T) {
	t.Parallel()

	cart := aggregate.NewCart(1)
	repo := &stubCartRepository{cart: cart}
	service := NewCartAppService(repo)

	require.NoError(t, service.CheckAll(context.Background(), &dto.CheckAllReq{UserID: 1}))
	require.NoError(t, service.UncheckAll(context.Background(), &dto.UncheckAllReq{UserID: 1}))
	require.NoError(t, service.Clear(context.Background(), &dto.ClearReq{UserID: 1}))
}

type stubCartRepository struct {
	cart      *aggregate.Cart
	savedCart *aggregate.Cart
}

func (s *stubCartRepository) GetByUserID(context.Context, int64) (*aggregate.Cart, error) {
	return s.cart, nil
}

func (s *stubCartRepository) Save(_ context.Context, cart *aggregate.Cart) error {
	s.savedCart = cart
	s.cart = cart
	return nil
}

func (s *stubCartRepository) AddItem(context.Context, int64, *aggregate.Cart) error {
	return nil
}

func (s *stubCartRepository) UpdateItemQuantity(context.Context, int64, int64, int32) error {
	return nil
}

func (s *stubCartRepository) RemoveItem(context.Context, int64, int64) error {
	return nil
}

func (s *stubCartRepository) Clear(context.Context, int64) error {
	return nil
}

var _ repository.CartRepository = (*stubCartRepository)(nil)
