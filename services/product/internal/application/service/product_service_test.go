package service

import (
	"context"
	"errors"
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

func newTestProduct(t *testing.T, stockValue int64) *aggregate.Product {
	t.Helper()

	price, err := valueobject.NewPrice(1000)
	require.NoError(t, err)
	stock, err := valueobject.NewStock(stockValue)
	require.NoError(t, err)
	category, err := valueobject.NewCategory(1, "phone")
	require.NoError(t, err)

	product := aggregate.NewProduct("product", "desc", "pic", price, stock, []valueobject.Category{category}, "thumb")
	product.ID = 1
	product.Sold = 3
	return product
}

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

func TestProductAppServiceQueryBehaviors(t *testing.T) {
	t.Parallel()

	product := newTestProduct(t, 8)
	repo := &stubProductRepository{
		product:      product,
		listProducts: []*aggregate.Product{product},
		total:        1,
	}
	service := NewProductAppService(repo, &stubProductEventPublisher{})

	getResp, err := service.GetProductByID(context.Background(), &dto.GetProductRequest{ID: uint32(product.ID)})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, getResp.StatusCode)
	require.Equal(t, "product", getResp.Product.Name)
	require.Equal(t, []string{"phone"}, getResp.Product.Categories)
	require.Equal(t, int64(8), getResp.Product.Stock)
	require.Equal(t, int64(3), getResp.Product.Sold)
	require.NotEmpty(t, getResp.Product.CreatedAt)
	require.NotEmpty(t, getResp.Product.UpdatedAt)
	require.Equal(t, "thumb", getResp.Product.ThumbnailURL)

	listResp, err := service.ListProducts(context.Background(), &dto.ListProductsRequest{Limit: 1})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, listResp.StatusCode)
	require.Len(t, listResp.Products, 1)
	require.False(t, listResp.HasMore)
	require.Zero(t, listResp.NextCursor)

	allResp, err := service.GetAllProducts(context.Background(), &dto.GetAllProductsRequest{Page: 2, PageSize: 10})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, allResp.StatusCode)
	require.Equal(t, int64(1), allResp.Total)
	require.Len(t, allResp.Products, 1)
	require.Equal(t, int64(2), allResp.Page)
	require.Equal(t, int64(10), allResp.PageSize)

	existResp, err := service.IsProductExist(context.Background(), &dto.IsExistProductRequest{ID: product.ID})
	require.NoError(t, err)
	require.True(t, existResp.Exist)
}

func TestProductAppServiceQueryErrorPaths(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("repo failed")
	service := NewProductAppService(&stubProductRepository{
		getErr:  expectedErr,
		listErr: expectedErr,
	}, &stubProductEventPublisher{})

	getResp, err := service.GetProductByID(context.Background(), &dto.GetProductRequest{ID: 1})
	require.ErrorIs(t, err, expectedErr)
	require.EqualValues(t, code.ProductNotFound, getResp.StatusCode)

	listResp, err := service.ListProducts(context.Background(), &dto.ListProductsRequest{Limit: 1})
	require.ErrorIs(t, err, expectedErr)
	require.EqualValues(t, code.ServerError, listResp.StatusCode)

	allResp, err := service.GetAllProducts(context.Background(), &dto.GetAllProductsRequest{Page: 1, PageSize: 10})
	require.ErrorIs(t, err, expectedErr)
	require.EqualValues(t, code.ServerError, allResp.StatusCode)

	existResp, err := service.IsProductExist(context.Background(), &dto.IsExistProductRequest{ID: 1})
	require.NoError(t, err)
	require.False(t, existResp.Exist)
}

func TestProductAppServiceDecreaseStock(t *testing.T) {
	t.Parallel()

	repo := &stubProductRepository{product: newTestProduct(t, 5)}
	publisher := &stubProductEventPublisher{}
	service := NewProductAppService(repo, publisher)

	err := service.DecreaseStock(context.Background(), repo.product.ID, 2)
	require.NoError(t, err)
	require.Equal(t, int64(3), repo.product.Stock.Value())
	require.Equal(t, int64(2), repo.decreasedQuantity)
	require.Equal(t, repo.product.ID, repo.decreasedProductID)
	require.NotNil(t, publisher.stockChangedEvent)
	require.Equal(t, int64(5), publisher.stockChangedEvent.OldStock)
	require.Equal(t, int64(3), publisher.stockChangedEvent.NewStock)
	require.Equal(t, int64(-2), publisher.stockChangedEvent.ChangedAmount)
}

func TestProductAppServiceDecreaseStockRejectsErrors(t *testing.T) {
	t.Parallel()

	t.Run("repo get failure", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("missing")
		service := NewProductAppService(&stubProductRepository{getErr: expectedErr}, &stubProductEventPublisher{})
		err := service.DecreaseStock(context.Background(), 1, 1)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("insufficient stock", func(t *testing.T) {
		t.Parallel()

		repo := &stubProductRepository{product: newTestProduct(t, 1)}
		service := NewProductAppService(repo, &stubProductEventPublisher{})
		err := service.DecreaseStock(context.Background(), repo.product.ID, 2)
		require.ErrorIs(t, err, valueobject.ErrInsufficientStock)
	})

	t.Run("repo decrease failure", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("atomic decrease failed")
		repo := &stubProductRepository{
			product:     newTestProduct(t, 5),
			decreaseErr: expectedErr,
		}
		service := NewProductAppService(repo, &stubProductEventPublisher{})
		err := service.DecreaseStock(context.Background(), repo.product.ID, 1)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestValueObjectCategoryPreparation(t *testing.T) {
	t.Parallel()

	category, err := valueobject.NewCategory(1, "phone")
	require.NoError(t, err)
	require.Equal(t, "phone", category.Name)
}

type stubProductRepository struct {
	product            *aggregate.Product
	savedProduct       *aggregate.Product
	listProducts       []*aggregate.Product
	total              int64
	getErr             error
	listErr            error
	deleteErr          error
	decreaseErr        error
	decreasedProductID int64
	decreasedQuantity  int64
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
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.product, nil
}

func (s *stubProductRepository) List(context.Context, int, int, *int64, *string) ([]*aggregate.Product, int64, error) {
	if s.listErr != nil {
		return nil, 0, s.listErr
	}
	if s.listProducts != nil {
		return s.listProducts, s.total, nil
	}
	if s.product == nil {
		return nil, 0, nil
	}
	return []*aggregate.Product{s.product}, 1, nil
}

func (s *stubProductRepository) Delete(context.Context, int64) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	return nil
}

func (s *stubProductRepository) BatchGetByIDs(context.Context, []int64) ([]*aggregate.Product, error) {
	return nil, nil
}

func (s *stubProductRepository) DecreaseStock(_ context.Context, productID int64, quantity int64) error {
	if s.decreaseErr != nil {
		return s.decreaseErr
	}
	s.decreasedProductID = productID
	s.decreasedQuantity = quantity
	return nil
}

func (s *stubProductRepository) IncreaseStock(context.Context, int64, int64) error {
	return nil
}

type stubProductEventPublisher struct {
	stockChangedEvent *domainevent.ProductStockChangedEvent
}

func (s *stubProductEventPublisher) PublishProductCreated(*domainevent.ProductCreatedEvent) error {
	return nil
}

func (s *stubProductEventPublisher) PublishProductUpdated(*domainevent.ProductUpdatedEvent) error {
	return nil
}

func (s *stubProductEventPublisher) PublishProductStockChanged(evt *domainevent.ProductStockChangedEvent) error {
	s.stockChangedEvent = evt
	return nil
}

func (s *stubProductEventPublisher) PublishProductDeleted(*domainevent.ProductDeletedEvent) error {
	return nil
}

var _ repository.ProductRepository = (*stubProductRepository)(nil)
var _ appevent.ProductEventPublisher = (*stubProductEventPublisher)(nil)
