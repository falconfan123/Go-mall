package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PaymentsModel = (*customPaymentsModel)(nil)

// paidStatusPaidValue payments.status 的 PAID 枚举值（DB bigint；与 pb PaymentStatus_PAYMENT_STATUS_PAID 对应）
const paidStatusPaidValue = 2

type (
	// PaymentsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPaymentsModel.
	PaymentsModel interface {
		paymentsModel
		WithSession(session sqlx.Session) PaymentsModel
		UpdateInfoByOrderId(ctx context.Context, newData *Payments) error
		Count(ctx context.Context) (int64, error)
		FindPage(ctx context.Context, userId uint32, offset, limit int) ([]*Payments, error)
		FindOneByOrderId(ctx context.Context, orderID string) (*Payments, error)
		CheckExistByOrderId(ctx context.Context, orderID string) (bool, error)
		FindExpired(ctx context.Context, limit int) ([]*Payments, error)
		UpdateStatusToPaidWithSession(ctx context.Context, session sqlx.Session, paymentId, transactionId string, paidAmount, paidAt int64) (rowsAffected int64, err error)
	}

	customPaymentsModel struct {
		*defaultPaymentsModel
	}
)

func (m *customPaymentsModel) CheckExistByOrderId(ctx context.Context, orderID string) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE \"order_id\" = $1", m.table)
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
	query := fmt.Sprintf("select %s from %s where \"order_id\" = $1 limit 1", paymentsRows, m.table)
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

// UpdateStatusToPaidWithSession 带状态机门槛的条件更新（design 决策 4）：
// 仅当当前状态非 PAID 时翻转为 PAID（并发 webhook 下只有一次生效），返回影响行数。
// rowsAffected==1 表示本次翻转成功（调用方负责同事务写入 outbox 待办）。
func (m *customPaymentsModel) UpdateStatusToPaidWithSession(ctx context.Context, session sqlx.Session, paymentId, transactionId string, paidAmount, paidAt int64) (int64, error) {
	query := fmt.Sprintf(`UPDATE %s SET "transaction_id" = NULLIF($1, ''), "paid_amount" = $2, "paid_at" = $3,
		"status" = $4, "updated_at" = now()
		WHERE "payment_id" = $5 AND "status" <> $4`, m.table)
	res, err := session.ExecCtx(ctx, query, transactionId, paidAmount, paidAt, paidStatusPaidValue, paymentId)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
