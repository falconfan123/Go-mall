package persistence

import (
	"context"
	"database/sql"

	cartmodel "github.com/falconfan123/Go-mall/dal/model/cart"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/carts/internal/domain/valueobject"
)

// CartRepositoryImpl 购物车仓储实现
type CartRepositoryImpl struct {
	cartsModel cartmodel.CartsModel
}

// NewCartRepositoryImpl 创建购物车仓储实现
func NewCartRepositoryImpl(cartsModel cartmodel.CartsModel) repository.CartRepository {
	return &CartRepositoryImpl{
		cartsModel: cartsModel,
	}
}

// GetByUserID 根据用户ID查询购物车
func (r *CartRepositoryImpl) GetByUserID(ctx context.Context, userID int64) (*aggregate.Cart, error) {
	// 1. 查询用户所有购物车项
	items, err := r.cartsModel.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. 如果没有购物车项，返回nil（表示用户还没有购物车）
	if len(items) == 0 {
		return nil, nil
	}

	// 3. 重构Cart聚合
	cartAgg := aggregate.NewCart(userID)
	for _, item := range items {
		// 处理NULL值
		productID := item.ProductId.Int64
		quantity := item.Quantity.Int64
		checked := item.Checked.Int64 == 1

		// 创建数量值对象
		qty, err := valueobject.NewQuantity(int32(quantity))
		if err != nil {
			// 无效数量，跳过该项或者返回错误
			continue
		}

		// 创建购物车项实体
		// 注意：商品名称、图片、价格需要从商品服务获取，这里暂时留空或者后续补充
		// 实际项目中，购物车表通常会冗余存储这些信息，避免每次查询商品服务
		cartItem := entity.NewCartItem(
			productID,
			"", // 商品名称，实际应该从数据库读取或者商品服务获取
			"", // 商品图片
			0,  // 商品价格
			qty,
		)
		cartItem.ID = item.Id
		cartItem.Checked = checked

		// 添加到购物车
		cartAgg.Items = append(cartAgg.Items, cartItem)
	}

	return cartAgg, nil
}

// Save 保存购物车
func (r *CartRepositoryImpl) Save(ctx context.Context, cart *aggregate.Cart) error {
	// 1. 先删除用户所有现有购物车项
	// 注意：这里采用全量覆盖的方式，实际生产环境可以优化为增量更新
	for _, item := range cart.Items {
		// 检查商品是否已存在
		_, exists, err := r.cartsModel.CheckCartItemExists(ctx, int32(cart.UserID), int32(item.ProductID))
		if err != nil {
			return err
		}

		if exists {
			// 更新现有项
			// 先查询现有记录ID
			existingID, _, err := r.cartsModel.CheckCartItemExists(ctx, int32(cart.UserID), int32(item.ProductID))
			if err != nil {
				return err
			}

			// 准备更新数据
			cartData := &cartmodel.Carts{
				Id: int64(existingID),
				UserId: sql.NullInt64{
					Int64: cart.UserID,
					Valid: true,
				},
				ProductId: sql.NullInt64{
					Int64: item.ProductID,
					Valid: true,
				},
				Quantity: sql.NullInt64{
					Int64: int64(item.Quantity.Value()),
					Valid: true,
				},
				Checked: sql.NullInt64{
					Int64: 0,
					Valid: true,
				},
			}
			if item.Checked {
				cartData.Checked.Int64 = 1
			}

			if err := r.cartsModel.Update(ctx, cartData); err != nil {
				return err
			}
		} else {
			// 插入新项
			cartData := &cartmodel.Carts{
				UserId: sql.NullInt64{
					Int64: cart.UserID,
					Valid: true,
				},
				ProductId: sql.NullInt64{
					Int64: item.ProductID,
					Valid: true,
				},
				Quantity: sql.NullInt64{
					Int64: int64(item.Quantity.Value()),
					Valid: true,
				},
				Checked: sql.NullInt64{
					Int64: 0,
					Valid: true,
				},
			}
			if item.Checked {
				cartData.Checked.Int64 = 1
			}

			_, err := r.cartsModel.Insert(ctx, cartData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AddItem 添加购物车项
func (r *CartRepositoryImpl) AddItem(ctx context.Context, userID int64, item *aggregate.Cart) error {
	// 这个方法可以优化为单独添加一项，这里暂时复用Save方法
	return r.Save(ctx, item)
}

// UpdateItemQuantity 更新购物车项数量
func (r *CartRepositoryImpl) UpdateItemQuantity(ctx context.Context, userID int64, productID int64, quantity int32) error {
	// 查询现有记录
	existingID, exists, err := r.cartsModel.CheckCartItemExists(ctx, int32(userID), int32(productID))
	if err != nil {
		return err
	}
	if !exists {
		return aggregate.ErrItemNotFound
	}

	// 更新数量
	cartData := &cartmodel.Carts{
		Id: int64(existingID),
		Quantity: sql.NullInt64{
			Int64: int64(quantity),
			Valid: true,
		},
	}

	return r.cartsModel.Update(ctx, cartData)
}

// RemoveItem 删除购物车项
func (r *CartRepositoryImpl) RemoveItem(ctx context.Context, userID int64, productID int64) error {
	return r.cartsModel.DeleteCartItem(ctx, int32(userID), int32(productID))
}

// Clear 清空购物车
func (r *CartRepositoryImpl) Clear(ctx context.Context, userID int64) error {
	// 查询所有项并逐个删除
	items, err := r.cartsModel.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := r.cartsModel.Delete(ctx, item.Id); err != nil {
			return err
		}
	}

	return nil
}
