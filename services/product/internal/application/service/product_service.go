package service

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/product/internal/application/dto"
	appevent "github.com/falconfan123/Go-mall/services/product/internal/application/event"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/event"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/valueobject"
	"github.com/google/uuid"
)

// ProductAppService 商品应用服务，编排业务流程
type ProductAppService struct {
	productRepo repository.ProductRepository
	eventPub    appevent.ProductEventPublisher
}

// NewProductAppService 创建商品应用服务
func NewProductAppService(
	productRepo repository.ProductRepository,
	eventPub appevent.ProductEventPublisher,
) *ProductAppService {
	return &ProductAppService{
		productRepo: productRepo,
		eventPub:    eventPub,
	}
}

// CreateProduct 创建商品
func (s *ProductAppService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.CreateProductResponse, error) {
	// 1. 创建值对象
	price, err := valueobject.NewPrice(req.Price)
	if err != nil {
		return &dto.CreateProductResponse{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, err
	}

	stock, err := valueobject.NewStock(req.Stock)
	if err != nil {
		return &dto.CreateProductResponse{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, err
	}

	// 处理分类
	var categories []valueobject.Category
	for _, catName := range req.Categories {
		// 这里简化处理，实际应该根据分类名称查询分类ID
		cat, err := valueobject.NewCategory(0, catName)
		if err != nil {
			return &dto.CreateProductResponse{
				StatusCode: code.Fail,
				StatusMsg:  err.Error(),
			}, err
		}
		categories = append(categories, cat)
	}

	// 2. 创建商品聚合根
	product := aggregate.NewProduct(
		req.Name,
		req.Description,
		string(req.Picture),
		price,
		stock,
		categories,
		"", // 缩略图后续处理
	)

	// 3. 保存商品
	if err := s.productRepo.Save(ctx, product); err != nil {
		return &dto.CreateProductResponse{
			StatusCode: code.ServerError,
			StatusMsg:  "failed to save product: " + err.Error(),
		}, err
	}

	// 4. 发布领域事件
	evt := &event.ProductCreatedEvent{
		BaseEvent: event.BaseEvent{
			EventID:    uuid.New().String(),
			EventType:  "product.created",
			OccurredAt: time.Now(),
		},
		ProductID:   product.ID,
		ProductName: product.Name,
		Price:       product.Price.Value(),
		Stock:       product.Stock.Value(),
	}
	if err := s.eventPub.PublishProductCreated(evt); err != nil {
		// 日志记录错误，不影响主流程
		fmt.Printf("failed to publish product created event: %v\n", err)
	}

	// 5. 返回响应
	return &dto.CreateProductResponse{
		StatusCode: code.Success,
		StatusMsg:  "success",
		ProductID:  product.ID,
	}, nil
}

// GetProductByID 根据ID获取商品
func (s *ProductAppService) GetProductByID(ctx context.Context, req *dto.GetProductRequest) (*dto.GetProductResponse, error) {
	product, err := s.productRepo.GetByID(ctx, int64(req.ID))
	if err != nil {
		return &dto.GetProductResponse{
			StatusCode: code.ProductNotFound,
			StatusMsg:  "product not found: " + err.Error(),
		}, err
	}

	// 转换为DTO
	productDTO := s.convertProductToDTO(product)

	return &dto.GetProductResponse{
		StatusCode: code.Success,
		StatusMsg:  "success",
		Product:    productDTO,
	}, nil
}

// UpdateProduct 更新商品
func (s *ProductAppService) UpdateProduct(ctx context.Context, req *dto.UpdateProductRequest) (*dto.UpdateProductResponse, error) {
	// 1. 查询现有商品
	product, err := s.productRepo.GetByID(ctx, req.ID)
	if err != nil {
		return &dto.UpdateProductResponse{
			StatusCode: code.ProductNotFound,
			StatusMsg:  "product not found: " + err.Error(),
		}, err
	}

	// 2. 创建新的值对象
	newPrice, err := valueobject.NewPrice(req.Price)
	if err != nil {
		return &dto.UpdateProductResponse{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, err
	}

	// 处理分类
	var categories []valueobject.Category
	for _, catName := range req.Categories {
		cat, err := valueobject.NewCategory(0, catName)
		if err != nil {
			return &dto.UpdateProductResponse{
				StatusCode: code.Fail,
				StatusMsg:  err.Error(),
			}, err
		}
		categories = append(categories, cat)
	}

	// 3. 更新商品信息
	oldPrice := product.Price.Value()
	product.UpdateInfo(
		req.Name,
		req.Description,
		string(req.Picture),
		newPrice,
		categories,
		"", // 缩略图处理
	)

	// 4. 保存更新
	if err := s.productRepo.Save(ctx, product); err != nil {
		return &dto.UpdateProductResponse{
			StatusCode: code.ServerError,
			StatusMsg:  "failed to update product: " + err.Error(),
		}, err
	}

	// 5. 如果价格变化，发布更新事件
	if oldPrice != newPrice.Value() {
		evt := &event.ProductUpdatedEvent{
			BaseEvent: event.BaseEvent{
				EventID:    uuid.New().String(),
				EventType:  "product.updated",
				OccurredAt: time.Now(),
			},
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       newPrice.Value(),
		}
		if err := s.eventPub.PublishProductUpdated(evt); err != nil {
			fmt.Printf("failed to publish product updated event: %v\n", err)
		}
	}

	// 6. 返回响应
	return &dto.UpdateProductResponse{
		StatusCode: code.Success,
		StatusMsg:  "success",
		ID:         product.ID,
	}, nil
}

// DeleteProduct 删除商品
func (s *ProductAppService) DeleteProduct(ctx context.Context, req *dto.DeleteProductRequest) (*dto.DeleteProductResponse, error) {
	// 1. 检查商品是否存在
	_, err := s.productRepo.GetByID(ctx, req.ID)
	if err != nil {
		return &dto.DeleteProductResponse{
			StatusCode: code.ProductNotFound,
			StatusMsg:  "product not found: " + err.Error(),
		}, err
	}

	// 2. 删除商品
	if err := s.productRepo.Delete(ctx, req.ID); err != nil {
		return &dto.DeleteProductResponse{
			StatusCode: code.ServerError,
			StatusMsg:  "failed to delete product: " + err.Error(),
		}, err
	}

	// 3. 发布删除事件
	evt := &event.ProductDeletedEvent{
		BaseEvent: event.BaseEvent{
			EventID:    uuid.New().String(),
			EventType:  "product.deleted",
			OccurredAt: time.Now(),
		},
		ProductID: req.ID,
	}
	if err := s.eventPub.PublishProductDeleted(evt); err != nil {
		fmt.Printf("failed to publish product deleted event: %v\n", err)
	}

	// 4. 返回响应
	return &dto.DeleteProductResponse{
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// ListProducts 游标分页查询商品
func (s *ProductAppService) ListProducts(ctx context.Context, req *dto.ListProductsRequest) (*dto.ListProductsResponse, error) {
	// 这里简化实现，实际应该使用游标查询
	products, total, err := s.productRepo.List(ctx, 1, int(req.Limit), nil, nil)
	if err != nil {
		return &dto.ListProductsResponse{
			StatusCode: code.ServerError,
			StatusMsg:  "failed to list products: " + err.Error(),
		}, err
	}

	// 转换为DTO
	productDTOs := make([]*dto.ProductDTO, 0, len(products))
	for _, p := range products {
		productDTOs = append(productDTOs, s.convertProductToDTO(p))
	}

	// 计算下一个游标
	hasMore := int64(len(products)) == req.Limit && total > req.Limit
	nextCursor := int64(0)
	if hasMore && len(products) > 0 {
		nextCursor = products[len(products)-1].ID
	}

	return &dto.ListProductsResponse{
		StatusCode: code.Success,
		StatusMsg:  "success",
		Products:   productDTOs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// GetAllProducts 分页查询所有商品
func (s *ProductAppService) GetAllProducts(ctx context.Context, req *dto.GetAllProductsRequest) (*dto.GetAllProductsResponse, error) {
	products, total, err := s.productRepo.List(ctx, int(req.Page), int(req.PageSize), nil, nil)
	if err != nil {
		return &dto.GetAllProductsResponse{
			StatusCode: code.ServerError,
			StatusMsg:  "failed to get all products: " + err.Error(),
		}, err
	}

	// 转换为DTO
	productDTOs := make([]*dto.ProductDTO, 0, len(products))
	for _, p := range products {
		productDTOs = append(productDTOs, s.convertProductToDTO(p))
	}

	return &dto.GetAllProductsResponse{
		StatusCode: code.Success,
		StatusMsg:  "success",
		Total:      total,
		Products:   productDTOs,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// IsProductExist 检查商品是否存在
func (s *ProductAppService) IsProductExist(ctx context.Context, req *dto.IsExistProductRequest) (*dto.IsExistProductResponse, error) {
	_, err := s.productRepo.GetByID(ctx, req.ID)
	if err != nil {
		return &dto.IsExistProductResponse{
			StatusCode: code.Success,
			StatusMsg:  "success",
			Exist:      false,
		}, nil
	}

	return &dto.IsExistProductResponse{
		StatusCode: code.Success,
		StatusMsg:  "success",
		Exist:      true,
	}, nil
}

// DecreaseStock 扣减库存（供内部服务调用）
func (s *ProductAppService) DecreaseStock(ctx context.Context, productID int64, quantity int64) error {
	// 1. 查询商品
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// 2. 领域层检查库存是否足够
	oldStock := product.Stock.Value()
	if err := product.AdjustStock(-quantity); err != nil {
		return err
	}

	// 3. 原子扣减库存
	if err := s.productRepo.DecreaseStock(ctx, productID, quantity); err != nil {
		return err
	}

	// 4. 发布库存变化事件
	evt := &event.ProductStockChangedEvent{
		BaseEvent: event.BaseEvent{
			EventID:    uuid.New().String(),
			EventType:  "product.stock_changed",
			OccurredAt: time.Now(),
		},
		ProductID:     productID,
		OldStock:      oldStock,
		NewStock:      product.Stock.Value(),
		ChangedAmount: -quantity,
	}
	if err := s.eventPub.PublishProductStockChanged(evt); err != nil {
		fmt.Printf("failed to publish stock changed event: %v\n", err)
	}

	return nil
}

// convertProductToDTO 将商品聚合根转换为DTO
func (s *ProductAppService) convertProductToDTO(product *aggregate.Product) *dto.ProductDTO {
	// 转换分类为字符串数组
	categories := make([]string, 0, len(product.Categories))
	for _, cat := range product.Categories {
		categories = append(categories, cat.Name)
	}

	return &dto.ProductDTO{
		ID:           uint32(product.ID),
		Name:         product.Name,
		Description:  product.Description,
		Picture:      product.Picture,
		Price:        product.Price.Value(),
		Stock:        product.Stock.Value(),
		Sold:         product.Sold,
		Categories:   categories,
		CreatedAt:    product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    product.UpdatedAt.Format(time.RFC3339),
		ThumbnailURL: product.ThumbnailURL,
	}
}
