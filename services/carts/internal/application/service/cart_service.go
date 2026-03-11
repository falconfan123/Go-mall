package service

import (
	"context"

	"github.com/falconfan123/Go-mall/services/carts/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/valueobject"
)

// CartAppService 购物车应用服务
type CartAppService struct {
	cartRepo repository.CartRepository
}

// NewCartAppService 创建购物车应用服务
func NewCartAppService(cartRepo repository.CartRepository) *CartAppService {
	return &CartAppService{
		cartRepo: cartRepo,
	}
}

// AddItem 添加商品到购物车
func (s *CartAppService) AddItem(ctx context.Context, req *dto.AddItemReq) error {
	// 1. 获取或创建购物车
	cart, err := s.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if cart == nil {
		cart = aggregate.NewCart(req.UserID)
	}

	// 2. 创建数量值对象
	qty, err := valueobject.NewQuantity(req.Quantity)
	if err != nil {
		return err
	}

	// 3. 创建购物车项实体
	item := entity.NewCartItem(
		req.ProductID,
		req.ProductName,
		req.ProductImage,
		req.ProductPrice,
		qty,
	)

	// 4. 调用领域逻辑添加商品
	if err := cart.AddItem(item); err != nil {
		return err
	}

	// 5. 保存购物车
	return s.cartRepo.Save(ctx, cart)
}

// UpdateItemQuantity 更新购物车商品数量
func (s *CartAppService) UpdateItemQuantity(ctx context.Context, req *dto.UpdateItemQuantityReq) error {
	// 1. 获取购物车
	cart, err := s.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if cart == nil {
		return aggregate.ErrItemNotFound
	}

	// 2. 获取当前数量
	currentQty, err := cart.GetItemQuantity(req.ProductID)
	if err != nil {
		return err
	}

	// 3. 计算增量
	delta := req.Quantity - currentQty
	if delta == 0 {
		return nil
	}

	// 4. 调用领域逻辑更新数量
	if delta > 0 {
		if err := cart.IncreaseItemQuantity(req.ProductID, delta); err != nil {
			return err
		}
	} else {
		if err := cart.DecreaseItemQuantity(req.ProductID, -delta); err != nil {
			return err
		}
	}

	// 5. 保存购物车
	return s.cartRepo.Save(ctx, cart)
}

// RemoveItem 删除购物车商品
func (s *CartAppService) RemoveItem(ctx context.Context, req *dto.RemoveItemReq) error {
	// 1. 获取购物车
	cart, err := s.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil
	}

	// 2. 调用领域逻辑删除商品
	if err := cart.RemoveItem(req.ProductID); err != nil {
		return err
	}

	// 3. 保存购物车
	return s.cartRepo.Save(ctx, cart)
}

// ToggleItemCheck 切换购物车商品选中状态
func (s *CartAppService) ToggleItemCheck(ctx context.Context, req *dto.ToggleItemCheckReq) error {
	// 1. 获取购物车
	cart, err := s.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if cart == nil {
		return aggregate.ErrItemNotFound
	}

	// 2. 调用领域逻辑切换选中状态
	if err := cart.ToggleItemCheck(req.ProductID); err != nil {
		return err
	}

	// 3. 保存购物车
	return s.cartRepo.Save(ctx, cart)
}

// CheckAll 全选购物车商品
func (s *CartAppService) CheckAll(ctx context.Context, req *dto.CheckAllReq) error {
	// 1. 获取购物车
	cart, err := s.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil
	}

	// 2. 调用领域逻辑全选
	cart.CheckAll()

	// 3. 保存购物车
	return s.cartRepo.Save(ctx, cart)
}

// UncheckAll 取消全选购物车商品
func (s *CartAppService) UncheckAll(ctx context.Context, req *dto.UncheckAllReq) error {
	// 1. 获取购物车
	cart, err := s.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil
	}

	// 2. 调用领域逻辑取消全选
	cart.UncheckAll()

	// 3. 保存购物车
	return s.cartRepo.Save(ctx, cart)
}

// Clear 清空购物车
func (s *CartAppService) Clear(ctx context.Context, req *dto.ClearReq) error {
	// 1. 获取购物车
	cart, err := s.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil
	}

	// 2. 调用领域逻辑清空
	cart.Clear()

	// 3. 保存购物车
	return s.cartRepo.Save(ctx, cart)
}

// GetCart 查询购物车
func (s *CartAppService) GetCart(ctx context.Context, req *dto.GetCartReq) (*dto.CartDTO, error) {
	// 1. 获取购物车
	cart, err := s.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		cart = aggregate.NewCart(req.UserID)
	}

	// 2. 转换为DTO
	items := make([]*dto.CartItemDTO, 0, len(cart.Items))
	var checkedCount int32
	var checkedAmount float64

	for _, item := range cart.Items {
		items = append(items, &dto.CartItemDTO{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			ProductPrice: item.ProductPrice,
			Quantity:     item.Quantity.Value(),
			Checked:      item.Checked,
		})

		if item.Checked {
			checkedCount++
			checkedAmount += item.ProductPrice * float64(item.Quantity.Value())
		}
	}

	return &dto.CartDTO{
		UserID:        cart.UserID,
		Items:         items,
		TotalQuantity: cart.GetTotalQuantity(),
		TotalAmount:   cart.GetTotalAmount(),
		CheckedCount:  checkedCount,
		CheckedAmount: checkedAmount,
	}, nil
}
