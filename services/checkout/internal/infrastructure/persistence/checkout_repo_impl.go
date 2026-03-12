package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	checkoutmodel "github.com/falconfan123/Go-mall/dal/model/checkout"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/repository"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CheckoutRepositoryImpl 预订单仓储实现
type CheckoutRepositoryImpl struct {
	checkoutModel      checkoutmodel.CheckoutsModel
	checkoutItemsModel checkoutmodel.CheckoutItemsModel
	conn               sqlx.SqlConn
}

// NewCheckoutRepositoryImpl 创建预订单仓储实现
func NewCheckoutRepositoryImpl(conn sqlx.SqlConn) repository.CheckoutRepository {
	return &CheckoutRepositoryImpl{
		checkoutModel:      checkoutmodel.NewCheckoutsModel(conn),
		checkoutItemsModel: checkoutmodel.NewCheckoutItemsModel(conn),
		conn:               conn,
	}
}

// GetByID 根据预订单ID查询
func (r *CheckoutRepositoryImpl) GetByID(ctx context.Context, preOrderID string) (*entity.Checkout, error) {
	// 查询预订单主表
	checkoutData, err := r.checkoutModel.FindOne(ctx, preOrderID)
	if err != nil {
		return nil, err
	}

	// 查询预订单商品项
	items, err := r.getItemsByPreOrderID(ctx, preOrderID)
	if err != nil {
		return nil, err
	}

	return r.convertToDomain(checkoutData, items), nil
}

// GetByUserID 根据用户ID查询
func (r *CheckoutRepositoryImpl) GetByUserID(ctx context.Context, userID int64) ([]*entity.Checkout, error) {
	// 简化实现，实际需要根据用户ID查询
	return []*entity.Checkout{}, nil
}

// Save 保存预订单
func (r *CheckoutRepositoryImpl) Save(ctx context.Context, checkout *entity.Checkout) error {
	checkoutData := r.convertToData(checkout)
	_, err := r.checkoutModel.Insert(ctx, checkoutData)
	if err != nil {
		return err
	}

	// 保存商品项
	for _, item := range checkout.Items {
		itemData := r.convertItemToData(checkout.PreOrderID, item)
		_, err := r.checkoutItemsModel.Insert(ctx, itemData)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update 更新预订单
func (r *CheckoutRepositoryImpl) Update(ctx context.Context, checkout *entity.Checkout) error {
	checkoutData := r.convertToData(checkout)
	return r.checkoutModel.Update(ctx, checkoutData)
}

// Delete 删除预订单
func (r *CheckoutRepositoryImpl) Delete(ctx context.Context, preOrderID string) error {
	// 先删除商品项
	err := r.deleteItemsByPreOrderID(ctx, preOrderID)
	if err != nil {
		return err
	}
	return r.checkoutModel.Delete(ctx, preOrderID)
}

// ListByUserID 查询用户的预订单列表
func (r *CheckoutRepositoryImpl) ListByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*entity.Checkout, int64, error) {
	// 简化实现，实际需要分页查询
	return []*entity.Checkout{}, 0, nil
}

// FindExpired 查找已过期的预订单
func (r *CheckoutRepositoryImpl) FindExpired(ctx context.Context, limit int) ([]*entity.Checkout, error) {
	// 简化实现，实际需要查询过期预订单
	return []*entity.Checkout{}, nil
}

// DecreaseStock 原子扣减库存
func (r *CheckoutRepositoryImpl) DecreaseStock(ctx context.Context, items []*entity.CheckoutItem) error {
	// 简化实现，实际需要调用库存服务扣减
	return nil
}

// IncreaseStock 原子恢复库存
func (r *CheckoutRepositoryImpl) IncreaseStock(ctx context.Context, items []*entity.CheckoutItem) error {
	// 简化实现，实际需要调用库存服务恢复
	return nil
}

// getItemsByPreOrderID 根据预订单ID查询商品项
func (r *CheckoutRepositoryImpl) getItemsByPreOrderID(ctx context.Context, preOrderID string) ([]*entity.CheckoutItem, error) {
	query := "SELECT id, pre_order_id, product_id, quantity, price, snapshot, created_at FROM checkout_items WHERE pre_order_id = ?"
	var items []checkoutmodel.CheckoutItems
	err := r.conn.QueryRowsCtx(ctx, &items, query, preOrderID)
	if err != nil {
		return nil, err
	}

	// 转换快照
	result := make([]*entity.CheckoutItem, 0, len(items))
	for _, item := range items {
		domainItem := &entity.CheckoutItem{
			ID:         item.Id,
			PreOrderID: item.PreOrderId,
			ProductID:  int64(item.ProductId),
			Quantity:   int(item.Quantity),
			Price:      item.Price,
			CreatedAt:  item.CreatedAt,
		}

		// 解析快照
		if item.Snapshot != "" {
			var snapshot entity.ProductSnapshot
			if err := json.Unmarshal([]byte(item.Snapshot), &snapshot); err == nil {
				domainItem.Snapshot = &snapshot
			}
		}

		result = append(result, domainItem)
	}

	return result, nil
}

// deleteItemsByPreOrderID 根据预订单ID删除商品项
func (r *CheckoutRepositoryImpl) deleteItemsByPreOrderID(ctx context.Context, preOrderID string) error {
	query := "DELETE FROM checkout_items WHERE pre_order_id = ?"
	_, err := r.conn.ExecCtx(ctx, query, preOrderID)
	return err
}

// convertToDomain 将数据模型转换为领域模型
func (r *CheckoutRepositoryImpl) convertToDomain(data *checkoutmodel.Checkouts, items []*entity.CheckoutItem) *entity.Checkout {
	// 转换过期时间
	expireTime := time.Unix(data.ExpireTime, 0)

	// 转换优惠券ID
	var couponIDs []string
	if data.CouponId.Valid && data.CouponId.String != "" {
		couponIDs = []string{data.CouponId.String}
	}

	return &entity.Checkout{
		PreOrderID:     data.PreOrderId,
		UserID:         int64(data.UserId),
		AddressID:      int64(data.AddressId),
		CouponIDs:      couponIDs,
		OriginalAmount: data.OriginalAmount,
		FinalAmount:    data.FinalAmount,
		Status:         entity.CheckoutStatus(data.Status),
		ExpireTime:     expireTime,
		Items:          items,
		CreatedAt:      data.CreatedAt,
		UpdatedAt:      data.UpdatedAt,
	}
}

// convertToData 将领域模型转换为数据模型
func (r *CheckoutRepositoryImpl) convertToData(checkout *entity.Checkout) *checkoutmodel.Checkouts {
	// 转换过期时间
	expireTime := pb.ExpireTime.Unix()

	// 转换优惠券ID
	var couponID sql.NullString
	if len(checkout.CouponIDs) > 0 {
		couponID = sql.NullString{
			String: checkout.CouponIDs[0],
			Valid:  true,
		}
	}

	return &checkoutmodel.Checkouts{
		PreOrderId:     checkout.PreOrderID,
		UserId:         uint64(checkout.UserID),
		AddressId:      uint64(checkout.AddressID),
		CouponId:       couponID,
		OriginalAmount: checkout.OriginalAmount,
		FinalAmount:    checkout.FinalAmount,
		Status:         int64(checkout.Status),
		ExpireTime:     expireTime,
		CreatedAt:      checkout.CreatedAt,
		UpdatedAt:      checkout.UpdatedAt,
	}
}

// convertItemToData 将领域模型商品项转换为数据模型
func (r *CheckoutRepositoryImpl) convertItemToData(preOrderID string, item *entity.CheckoutItem) *checkoutmodel.CheckoutItems {
	var snapshot string
	if item.Snapshot != nil {
		snapshotBytes, _ := json.Marshal(item.Snapshot)
		snapshot = string(snapshotBytes)
	}

	return &checkoutmodel.CheckoutItems{
		PreOrderId: preOrderID,
		ProductId:  uint64(item.ProductID),
		Quantity:   uint64(item.Quantity),
		Price:      item.Price,
		Snapshot:   snapshot,
		CreatedAt:  item.CreatedAt,
	}
}
