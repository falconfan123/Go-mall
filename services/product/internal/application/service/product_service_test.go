package service

import (
	"context"
	"testing"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/product/internal/application/dto"
	appevent "github.com/falconfan123/Go-mall/services/product/internal/application/event"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/aggregate"
	domainevent "github.com/falconfan123/Go-mall/services/product/internal/domain/event"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestProductAppServiceCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	repo := &stubProductRepository{}
	publisher := &stubProductEventPublisher{}
	service := NewProductAppService(repo, publisher)

	createResp, err := service.CreateProduct(context.Background(), &dto.CreateProductRequest{
		Name:        "product",
		Description: "desc",
		Price:       1000,
		Stock:       5,
		Picture:     []byte("img"),
		Categories:  []string{"phone"},
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, createResp.StatusCode)
	require.NotNil(t, repo.savedProduct)

	repo.product = repo.savedProduct
	updateResp, err := service.UpdateProduct(context.Background(), &dto.UpdateProductRequest{
		ID:          repo.product.ID,
		Name:        "product-v2",
		Description: "desc2",
		Price:       1200,
		Picture:     []byte("img2"),
		Categories:  []string{"phone"},
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, updateResp.StatusCode)

	deleteResp, err := service.DeleteProduct(context.Background(), &dto.DeleteProductRequest{ID: repo.product.ID})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, deleteResp.StatusCode)
}

type stubProductRepository struct {
	product      *aggregate.Product
	savedProduct *aggregate.Product
}

func (s *stubProductRepository) Save(_ context.Context, product *aggregate.Product) error {
	if product.ID == 0 {
		product.ID = 1
	}
	s.savedProduct = product
	s.product = product
	return nil
}

func (s *stubProductRepository) GetByID(context.Context, int64) (*aggregate.Product, error) {
	return s.product, nil
}

func (s *stubProductRepository) List(context.Context, int, int, *int64, *string) ([]*aggregate.Product, int64, error) {
	if s.product == nil {
		return nil, 0, nil
	}
	return []*aggregate.Product{s.product}, 1, nil
}

func (s *stubProductRepository) Delete(context.Context, int64) error {
	return nil
}

func (s *stubProductRepository) BatchGetByIDs(context.Context, []int64) ([]*aggregate.Product, error) {
	return nil, nil
}

func (s *stubProductRepository) DecreaseStock(context.Context, int64, int64) error {
	return nil
}

func (s *stubProductRepository) IncreaseStock(context.Context, int64, int64) error {
	return nil
}

type stubProductEventPublisher struct{}

func (s *stubProductEventPublisher) PublishProductCreated(*domainevent.ProductCreatedEvent) error {
	return nil
}

func (s *stubProductEventPublisher) PublishProductUpdated(*domainevent.ProductUpdatedEvent) error {
	return nil
}

func (s *stubProductEventPublisher) PublishProductStockChanged(*domainevent.ProductStockChangedEvent) error {
	return nil
}

func (s *stubProductEventPublisher) PublishProductDeleted(*domainevent.ProductDeletedEvent) error {
	return nil
}

var _ repository.ProductRepository = (*stubProductRepository)(nil)
var _ appevent.ProductEventPublisher = (*stubProductEventPublisher)(nil)

func TestValueObjectCategoryPreparation(t *testing.T) {
	t.Parallel()

	category, err := valueobject.NewCategory(1, "phone")
	require.NoError(t, err)
	require.Equal(t, "phone", category.Name)
}
