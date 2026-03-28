package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PaymentsModel = (*customPaymentsModel)(nil)

type (
	// PaymentsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPaymentsModel.
	PaymentsModel interface {
		paymentsModel
		WithSession(session sqlx.Session) PaymentsModel
		UpdateInfoByOrderId(ctx context.Context, newData *Payments) error
		Count(ctx context.Context) (int64, error)
		FindPage(ctx context.Context, userId uint32, offset, limit int) ([]*Payments, error)
		FindOneByOrderId(ctx context.Context, pre_order_id string) (*Payments, error)
		CheckExistByOrderId(ctx context.Context, orderID string) (bool, error)
		FindExpired(ctx context.Context, limit int) ([]*Payments, error)
	}

	customPaymentsModel struct {
		*defaultPaymentsModel
	}
)

func (m *customPaymentsModel) CheckExistByOrderId(ctx context.Context, orderID string) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE \"pre_order_id\" = $1", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, orderID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// NewPaymentsModel returns a model for the database table.
func NewPaymentsModel(conn sqlx.SqlConn) PaymentsModel {
	return &customPaymentsModel{
		defaultPaymentsModel: newPaymentsModel(conn),
	}
}

func (m *customPaymentsModel) WithSession(session sqlx.Session) PaymentsModel {
	return NewPaymentsModel(sqlx.NewSqlConnFromSession(session))
}
func (m *defaultPaymentsModel) UpdateInfoByOrderId(ctx context.Context, newData *Payments) error {
	// PostgreSQL $1, $2, ... placeholders
	paymentsRowsWithHolder := "\"transaction_id\"=$1, \"status\"=$2, \"paid_at\"=$3"

	// 构造 SQL 更新语句
	query := fmt.Sprintf("update %s set %s where \"order_id\" = $4", m.table, paymentsRowsWithHolder)

	// 执行更新操作
	_, err := m.conn.ExecCtx(ctx, query,
		newData.TransactionId,
		newData.Status,
		newData.PaidAt,
		newData.OrderId,
	)
	return err
}

// 查询支付记录
func (m *defaultPaymentsModel) FindPage(ctx context.Context, userId uint32, offset, limit int) ([]*Payments, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE \"user_id\" = $1 LIMIT $2 OFFSET $3", m.table)
	var payments []*Payments
	err := m.conn.QueryRowsCtx(ctx, &payments, query, userId, limit, offset)
	if err != nil {
		return nil, err
	}
	return payments, nil
}
func (m *defaultPaymentsModel) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (m *defaultPaymentsModel) FindOneByOrderId(ctx context.Context, orderID string) (*Payments, error) {
	// PostgreSQL doesn't support "LIMIT n FOR SHARE", using "LIMIT n"
	// 注意：这里实际查询的是 pre_order_id 字段
	query := fmt.Sprintf("select %s from %s where \"pre_order_id\" = $1 limit 1", paymentsRows, m.table)
	var resp Payments
	err := m.conn.QueryRowCtx(ctx, &resp, query, orderID)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindExpired 查找已过期的支付单
func (m *defaultPaymentsModel) FindExpired(ctx context.Context, limit int) ([]*Payments, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE \"status\" = $1 AND \"expire_time\" < $2 LIMIT $3", paymentsRows, m.table)
	var payments []*Payments
	err := m.conn.QueryRowsCtx(ctx, &payments, query, 1, time.Now().Unix(), limit)
	if err != nil {
		return nil, err
	}
	return payments, nil
}
