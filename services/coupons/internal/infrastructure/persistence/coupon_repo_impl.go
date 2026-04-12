package persistence

import (
	"context"
	"time"

	couponmodel "github.com/falconfan123/Go-mall/dal/model/coupons/coupon"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/valueobject"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CouponRepositoryImpl 优惠券仓储实现
type CouponRepositoryImpl struct {
	couponModel couponmodel.CouponsModel
	conn        sqlx.SqlConn
}

// NewCouponRepositoryImpl 创建优惠券仓储实现
func NewCouponRepositoryImpl(conn sqlx.SqlConn) repository.CouponRepository {
	return &CouponRepositoryImpl{
		couponModel: couponmodel.NewCouponsModel(conn),
		conn:        conn,
	}
}

// GetByID 根据ID查询优惠券
func (r *CouponRepositoryImpl) GetByID(ctx context.Context, id string) (*aggregate.Coupon, error) {
	couponData, err := r.couponModel.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertToDomain(couponData)
}

// Save 保存优惠券
func (r *CouponRepositoryImpl) Save(ctx context.Context, coupon *aggregate.Coupon) error {
	couponData := r.convertToData(coupon)
	_, err := r.couponModel.Insert(ctx, couponData)
	return err
}

// Update 更新优惠券
func (r *CouponRepositoryImpl) Update(ctx context.Context, coupon *aggregate.Coupon) error {
	couponData := r.convertToData(coupon)
	return r.couponModel.Update(ctx, couponData)
}

// Delete 删除优惠券
func (r *CouponRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.couponModel.Delete(ctx, id)
}

// List 查询优惠券列表
func (r *CouponRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*aggregate.Coupon, int64, error) {
	// 简化实现，实际需要分页查询
	// 这里返回空列表
	return []*aggregate.Coupon{}, 0, nil
}

// ListAvailable 查询可用优惠券列表
func (r *CouponRepositoryImpl) ListAvailable(ctx context.Context, page, pageSize int) ([]*aggregate.Coupon, int64, error) {
	// 简化实现，实际需要根据状态和时间查询可用优惠券
	return []*aggregate.Coupon{}, 0, nil
}

// DecreaseStock 原子扣减库存
func (r *CouponRepositoryImpl) DecreaseStock(ctx context.Context, couponID string, count int) error {
	query := "UPDATE coupons SET remaining_count = remaining_count - ? WHERE id = ? AND remaining_count >= ?"
	result, err := r.conn.ExecCtx(ctx, query, count, couponID, count)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return aggregate.ErrCouponOutOfStock
	}
	return nil
}

// IncreaseStock 原子增加库存
func (r *CouponRepositoryImpl) IncreaseStock(ctx context.Context, couponID string, count int) error {
	query := "UPDATE coupons SET remaining_count = remaining_count + ? WHERE id = ?"
	_, err := r.conn.ExecCtx(ctx, query, count, couponID)
	return err
}

// convertToDomain 将数据模型转换为领域模型
func (r *CouponRepositoryImpl) convertToDomain(data *couponmodel.Coupons) (*aggregate.Coupon, error) {
	// 转换折扣信息
	discount, err := valueobject.NewDiscount(
		valueobject.CouponType(data.Type),
		data.Value,
		data.MinAmount,
	)
	if err != nil {
		return nil, err
	}

	// 转换有效期
	validPeriod, err := valueobject.NewValidPeriod(data.StartTime, data.EndTime)
	if err != nil {
		return nil, err
	}

	// 转换状态
	status := valueobject.CouponStatus(data.Status)

	return &aggregate.Coupon{
		ID:             data.Id,
		Name:           data.Name,
		Discount:       discount,
		ValidPeriod:    validPeriod,
		Status:         status,
		TotalCount:     data.TotalCount,
		RemainingCount: data.RemainingCount,
		CreatedAt:      data.CreatedAt,
		UpdatedAt:      data.UpdatedAt,
	}, nil
}

// convertToData 将领域模型转换为数据模型
func (r *CouponRepositoryImpl) convertToData(coupon *aggregate.Coupon) *couponmodel.Coupons {
	return &couponmodel.Coupons{
		Id:             coupon.ID,
		Name:           coupon.Name,
		Type:           int64(coupon.Discount.CouponType()),
		Value:          coupon.Discount.Value(),
		MinAmount:      coupon.Discount.MinAmount(),
		StartTime:      coupon.ValidPeriod.StartTime(),
		EndTime:        coupon.ValidPeriod.EndTime(),
		Status:         int64(coupon.Status),
		TotalCount:     coupon.TotalCount,
		RemainingCount: coupon.RemainingCount,
		CreatedAt:      coupon.CreatedAt,
		UpdatedAt:      coupon.UpdatedAt,
	}
}

// PageCoupon 分页查询优惠券
type PageCoupon struct {
	Id             string
	Name           string
	Type           int64
	Value          int64
	MinAmount      int64
	StartTime      time.Time
	EndTime        time.Time
	Status         int64
	TotalCount     uint64
	RemainingCount uint64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ListByPage 分页查询优惠券
func (r *CouponRepositoryImpl) ListByPage(ctx context.Context, offset, limit int) ([]*PageCoupon, int64, error) {
	query := "SELECT id, name, type, value, min_amount, start_time, end_time, status, total_count, remaining_count, created_at, updated_at FROM coupons ORDER BY created_at DESC LIMIT ?, ?"
	var results []*PageCoupon
	err := r.conn.QueryRowsCtx(ctx, &results, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	var count int64
	countQuery := "SELECT COUNT(*) FROM coupons"
	err = r.conn.QueryRowCtx(ctx, &count, countQuery)
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}
