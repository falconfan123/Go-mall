package persistence

import (
	"context"
	"database/sql"
	"time"

	usercouponmodel "github.com/falconfan123/Go-mall/dal/model/coupons/user_coupons"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/repository"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserCouponRepositoryImpl 用户优惠券仓储实现
type UserCouponRepositoryImpl struct {
	userCouponModel usercouponmodel.UserCouponsModel
	conn            sqlx.SqlConn
}

// NewUserCouponRepositoryImpl 创建用户优惠券仓储实现
func NewUserCouponRepositoryImpl(conn sqlx.SqlConn) repository.UserCouponRepository {
	return &UserCouponRepositoryImpl{
		userCouponModel: usercouponmodel.NewUserCouponsModel(conn),
		conn:            conn,
	}
}

// GetByID 根据ID查询用户优惠券
func (r *UserCouponRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.UserCoupon, error) {
	data, err := r.userCouponModel.FindOne(ctx, uint64(id))
	if err != nil {
		return nil, err
	}
	return r.convertToDomain(data), nil
}

// GetByUserIDAndCouponID 根据用户ID和优惠券ID查询
func (r *UserCouponRepositoryImpl) GetByUserIDAndCouponID(ctx context.Context, userID int64, couponID string) (*entity.UserCoupon, error) {
	data, err := r.userCouponModel.FindOneByUserIdCouponId(ctx, uint64(userID), couponID)
	if err != nil {
		return nil, err
	}
	return r.convertToDomain(data), nil
}

// ListByUserID 查询用户的优惠券列表
func (r *UserCouponRepositoryImpl) ListByUserID(ctx context.Context, userID int64, status *entity.UserCouponStatus, page, pageSize int) ([]*entity.UserCoupon, int64, error) {
	// 简化实现，实际需要根据状态分页查询
	// 返回空列表
	return []*entity.UserCoupon{}, 0, nil
}

// Save 保存用户优惠券
func (r *UserCouponRepositoryImpl) Save(ctx context.Context, userCoupon *entity.UserCoupon) error {
	data := r.convertToData(userCoupon)
	_, err := r.userCouponModel.Insert(ctx, data)
	return err
}

// Update 更新用户优惠券
func (r *UserCouponRepositoryImpl) Update(ctx context.Context, userCoupon *entity.UserCoupon) error {
	data := r.convertToData(userCoupon)
	return r.userCouponModel.Update(ctx, data)
}

// Delete 删除用户优惠券
func (r *UserCouponRepositoryImpl) Delete(ctx context.Context, id int64) error {
	return r.userCouponModel.Delete(ctx, uint64(id))
}

// CountByUserIDAndCouponID 统计用户领取某优惠券的数量
func (r *UserCouponRepositoryImpl) CountByUserIDAndCouponID(ctx context.Context, userID int64, couponID string) (int64, error) {
	// 简化实现，需要自定义查询
	return 0, nil
}

// FindAvailableByUserID 查询用户可用的优惠券
func (r *UserCouponRepositoryImpl) FindAvailableByUserID(ctx context.Context, userID int64, orderAmount int64) ([]*entity.UserCoupon, error) {
	// 简化实现，实际需要根据状态和时间查询可用优惠券
	// 返回空列表
	return []*entity.UserCoupon{}, nil
}

// convertToDomain 将数据模型转换为领域模型
func (r *UserCouponRepositoryImpl) convertToDomain(data *usercouponmodel.UserCoupons) *entity.UserCoupon {
	uc := &entity.UserCoupon{
		ID:       int64(data.Id),
		UserID:   int64(data.UserId),
		CouponID: data.CouponId,
		Status:   entity.UserCouponStatus(data.Status),
		GetTime:  data.CreatedAt,
	}

	// 转换使用时间
	if data.UsedAt.Valid {
		uc.UseTime = &data.UsedAt.Time
	}

	// 转换订单ID
	if data.OrderId.Valid {
		uc.OrderID = &data.OrderId.String
	}

	return uc
}

// convertToData 将领域模型转换为数据模型
func (r *UserCouponRepositoryImpl) convertToData(userCoupon *entity.UserCoupon) *usercouponmodel.UserCoupons {
	data := &usercouponmodel.UserCoupons{
		Id:       uint64(userCoupon.ID),
		UserId:   uint64(userCoupon.UserID),
		CouponId: userCoupon.CouponID,
		Status:   int64(userCoupon.Status),
	}

	// 转换使用时间
	if userCoupon.UseTime != nil {
		data.UsedAt = sql.NullTime{
			Time:  *userCoupon.UseTime,
			Valid: true,
		}
	}

	// 转换订单ID
	if userCoupon.OrderID != nil {
		data.OrderId = sql.NullString{
			String: *userCoupon.OrderID,
			Valid:  true,
		}
	}

	return data
}

// ListUnusedByUserID 查询用户未使用的优惠券
func (r *UserCouponRepositoryImpl) ListUnusedByUserID(ctx context.Context, userID int64) ([]*entity.UserCoupon, error) {
	// 简化实现，实际需要根据状态查询
	return []*entity.UserCoupon{}, nil
}

// ListUsedByUserID 查询用户已使用的优惠券
func (r *UserCouponRepositoryImpl) ListUsedByUserID(ctx context.Context, userID int64) ([]*entity.UserCoupon, error) {
	// 简化实现，实际需要根据状态查询
	return []*entity.UserCoupon{}, nil
}

// ListExpiredByUserID 查询用户已过期的优惠券
func (r *UserCouponRepositoryImpl) ListExpiredByUserID(ctx context.Context, userID int64) ([]*entity.UserCoupon, error) {
	// 简化实现，实际需要根据状态和时间查询
	return []*entity.UserCoupon{}, nil
}

// MarkAsUsed 标记为已使用
func (r *UserCouponRepositoryImpl) MarkAsUsed(ctx context.Context, userCouponID int64, orderID string) error {
	userCoupon, err := r.GetByID(ctx, userCouponID)
	if err != nil {
		return err
	}

	now := time.Now()
	userCoupon.UseTime = &now
	userCoupon.OrderID = &orderID
	userCoupon.Status = entity.UserCouponStatusUsed

	return r.Update(ctx, userCoupon)
}

// CancelUse 取消使用
func (r *UserCouponRepositoryImpl) CancelUse(ctx context.Context, userCouponID int64) error {
	userCoupon, err := r.GetByID(ctx, userCouponID)
	if err != nil {
		return err
	}

	userCoupon.Status = entity.UserCouponStatusUnused
	userCoupon.UseTime = nil
	userCoupon.OrderID = nil

	return r.Update(ctx, userCoupon)
}
