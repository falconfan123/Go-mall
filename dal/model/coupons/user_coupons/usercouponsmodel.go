package user_coupons

import (
	"context"
	"errors"
	"fmt"
	coupons "github.com/falconfan123/Go-mall/common/types/coupons"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserCouponsModel = (*customUserCouponsModel)(nil)

type (
	// UserCouponsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserCouponsModel.
	UserCouponsModel interface {
		userCouponsModel
		WithSession(session sqlx.Session) UserCouponsModel
		QueryUserCoupons(ctx context.Context, userId, page, pageSize int32) ([]*UserCoupons, error)
		CheckUserCouponExistWithLock(ctx context.Context, session sqlx.Session, userId uint64, couponId string) (bool, error)
		GetUserCouponByUserIdCouponIdWithLock(ctx context.Context, session sqlx.Session, userId uint64, couponId string) (*UserCoupons, error)
		GetStatusByUserIdCouponId(ctx context.Context, userid int32, couponId string) (*Status, error)
		UpdateStatusOrderById(ctx context.Context, orderId string, id int, status coupons.CouponStatus) error
		LockUserCoupon(ctx context.Context, session sqlx.Session, uCouponID uint64) error
		CheckUserCouponStatus(ctx context.Context, session sqlx.Session, u uint64, id string) (int64, error)
	}

	customUserCouponsModel struct {
		*defaultUserCouponsModel
	}
)

func (m *customUserCouponsModel) CheckUserCouponStatus(ctx context.Context, session sqlx.Session, u uint64, id string) (int64, error) {
	var status int64
	query := fmt.Sprintf("select \"status\" from %s where \"user_id\" = $1 and \"coupon_id\" = $2 limit 1", m.table)
	err := session.QueryRowCtx(ctx, &status, query, u, id)
	switch {
	case err == nil:
		return status, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return 0, sqlx.ErrNotFound
	default:
		return 0, err
	}
}

func (m *customUserCouponsModel) LockUserCoupon(ctx context.Context, session sqlx.Session, uCouponID uint64) error {
	query := fmt.Sprintf("update %s set \"status\" = $1 where \"id\" = $2", m.table)
	_, err := session.ExecCtx(ctx, query, coupons.CouponStatusLocked, uCouponID)
	return err
}

func (m *customUserCouponsModel) UpdateStatusOrderById(ctx context.Context, orderId string, id int, used coupons.CouponStatus) error {
	query := fmt.Sprintf("update %s set \"status\" = $1, \"order_id\" = $2, used_at = now() where \"id\" = $3", m.table)
	_, err := m.conn.ExecCtx(ctx, query, used, orderId, id)
	return err
}

func (m *customUserCouponsModel) GetStatusByUserIdCouponId(ctx context.Context, userid int32, couponId string) (*Status, error) {
	var status Status
	query := fmt.Sprintf("select \"id\",\"status\" from %s where \"user_id\" = $1 and \"coupon_id\" = $2 limit 1", m.table)
	err := m.conn.QueryRowCtx(ctx, &status, query, userid, couponId)
	switch {
	case err == nil:
		return &status, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserCouponsModel) GetUserCouponByUserIdCouponIdWithLock(ctx context.Context, session sqlx.Session, userId uint64, couponId string) (*UserCoupons, error) {
	query := fmt.Sprintf("select %s from %s where \"user_id\" = $1 and \"coupon_id\" = $2 limit 1 for update", userCouponsRows, m.table)
	var resp UserCoupons
	err := session.QueryRowCtx(ctx, &resp, query, userId, couponId)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserCouponsModel) CheckUserCouponExistWithLock(ctx context.Context, session sqlx.Session, userId uint64, couponId string) (bool, error) {
	var id uint64
	query := fmt.Sprintf("select \"id\" from %s where \"user_id\" = $1 and \"coupon_id\" = $2 limit 1 for share", m.table)
	err := session.QueryRowCtx(ctx, &id, query, userId, couponId)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (m *customUserCouponsModel) QueryUserCoupons(ctx context.Context, userId, page, pageSize int32) ([]*UserCoupons, error) {
	query := fmt.Sprintf("select %s from %s where \"user_id\" = $1 order by \"created_at\" desc limit $2 offset $3", userCouponsRows, m.table)
	var resp []*UserCoupons
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, pageSize, (page-1)*pageSize)
	switch {
	case err == nil:
		return resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return resp, nil
	default:
		return nil, err
	}
}

// NewUserCouponsModel returns a model for the database table.
func NewUserCouponsModel(conn sqlx.SqlConn) UserCouponsModel {
	return &customUserCouponsModel{
		defaultUserCouponsModel: newUserCouponsModel(conn),
	}
}

func (m *customUserCouponsModel) WithSession(session sqlx.Session) UserCouponsModel {
	return NewUserCouponsModel(sqlx.NewSqlConnFromSession(session))
}
