package coupon_usage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CouponUsageModel = (*customCouponUsageModel)(nil)

type (
	// CouponUsageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCouponUsageModel.
	CouponUsageModel interface {
		couponUsageModel
		WithSession(session sqlx.Session) CouponUsageModel
		QueryUsageListByUserId(ctx context.Context, userId uint64, page, size int32) ([]*CouponUsage, error)
		DeleteByOrderId(ctx context.Context, session sqlx.Session, orderId string) error
	}

	customCouponUsageModel struct {
		*defaultCouponUsageModel
	}
)

func (m *customCouponUsageModel) QueryUsageListByUserId(ctx context.Context, userId uint64, page, size int32) ([]*CouponUsage, error) {

	offset := (page - 1) * size

	query := fmt.Sprintf("SELECT %s FROM %s WHERE \"user_id\" = $1 ORDER BY \"applied_at\" DESC LIMIT $2 OFFSET $3",
		"id, order_id, coupon_id, user_id, coupon_type, origin_value, discount_amount, applied_at", m.table)

	var resp []*CouponUsage
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, size, offset)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*CouponUsage, 0), nil
		}
		return nil, err
	}
	return resp, nil
}

// NewCouponUsageModel returns a model for the database table.
func NewCouponUsageModel(conn sqlx.SqlConn) CouponUsageModel {
	return &customCouponUsageModel{
		defaultCouponUsageModel: newCouponUsageModel(conn),
	}
}

func (m *customCouponUsageModel) WithSession(session sqlx.Session) CouponUsageModel {
	return NewCouponUsageModel(sqlx.NewSqlConnFromSession(session))
}

// DeleteByOrderId 按订单删除使用记录（Saga 补偿：回滚"已使用"时同步撤销流水）。
func (m *customCouponUsageModel) DeleteByOrderId(ctx context.Context, session sqlx.Session, orderId string) error {
	query := fmt.Sprintf("delete from %s where \"order_id\" = $1", m.table)
	_, err := session.ExecCtx(ctx, query, orderId)
	return err
}
