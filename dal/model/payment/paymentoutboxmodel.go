package payment

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// PaymentOutbox 结算待办（webhook 翻转事务内写入，relay 接力 EnsureSaga）。
type PaymentOutbox struct {
	Id          int64          `db:"id"`
	OrderId     string         `db:"order_id"`
	PaymentId   string         `db:"payment_id"`
	UserId      int64          `db:"user_id"`
	Payload     string         `db:"payload"` // SettlementInput 的 encoding/json 序列化（design 决策 3）
	Status      string         `db:"status"`  // pending|done|dead
	Attempts    int64          `db:"attempts"`
	NextRetryAt sql.NullTime   `db:"next_retry_at"`
	DoneAt      sql.NullTime   `db:"done_at"`
	LastError   sql.NullString `db:"last_error"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
}

var _ PaymentOutboxModel = (*customPaymentOutboxModel)(nil)

type (
	// PaymentOutboxModel 结算待办模型。
	PaymentOutboxModel interface {
		InsertPendingWithSession(ctx context.Context, session sqlx.Session, orderId, paymentId string, userId int64, payload string) error
		FetchPendingBatch(ctx context.Context, limit int) ([]*PaymentOutbox, error)
		ClaimPending(ctx context.Context, id int64, claimSeconds int) (bool, error)
		MarkDone(ctx context.Context, id int64) error
		MarkDead(ctx context.Context, id int64, lastErr string) error
		IncrAttempt(ctx context.Context, id int64, backoffSeconds int) (int64, error)
	}

	paymentOutboxModel struct {
		conn  sqlx.SqlConn
		table string
	}

	customPaymentOutboxModel struct {
		paymentOutboxModel
	}
)

func NewPaymentOutboxModel(conn sqlx.SqlConn) PaymentOutboxModel {
	return &customPaymentOutboxModel{
		paymentOutboxModel{conn: conn, table: "payment_outbox"},
	}
}

// InsertPendingWithSession 翻转事务内写入待办；order_id 唯一，并发重复插入静默忽略。
func (m *customPaymentOutboxModel) InsertPendingWithSession(ctx context.Context, session sqlx.Session, orderId, paymentId string, userId int64, payload string) error {
	query := fmt.Sprintf(`INSERT INTO %s (order_id, payment_id, user_id, payload, status)
		VALUES ($1, $2, $3, $4, 'pending')
		ON CONFLICT (order_id) DO NOTHING`, m.table)
	_, err := session.ExecCtx(ctx, query, orderId, paymentId, userId, payload)
	return err
}

// FetchPendingBatch 取回到期 pending 待办（不含 next_retry_at 未来的租约内行）。
func (m *customPaymentOutboxModel) FetchPendingBatch(ctx context.Context, limit int) ([]*PaymentOutbox, error) {
	query := fmt.Sprintf(`SELECT id, order_id, payment_id, user_id, payload, status, attempts,
		next_retry_at, done_at, last_error, created_at, updated_at
		FROM %s WHERE status = 'pending'
		  AND (next_retry_at IS NULL OR next_retry_at <= now())
		ORDER BY id LIMIT $1`, m.table)
	var out []*PaymentOutbox
	err := m.conn.QueryRowsCtx(ctx, &out, query, limit)
	return out, err
}

// ClaimPending 并发安全认领：条件更新推进租约，rows==1 表示本实例获得处理权
// （崩溃后租约到期自然重新可见，EnsureSaga 幂等收敛重复触发）。
func (m *customPaymentOutboxModel) ClaimPending(ctx context.Context, id int64, claimSeconds int) (bool, error) {
	query := fmt.Sprintf(`UPDATE %s SET next_retry_at = now() + ($2 || ' seconds')::interval,
		updated_at = now()
		WHERE id = $1 AND status = 'pending'`, m.table)
	res, err := m.conn.ExecCtx(ctx, query, id, fmt.Sprint(claimSeconds))
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()
	return rows == 1, nil
}

// MarkDone EnsureSaga 提交成功即 done（不等待 saga 终态，design 决策 5）。
func (m *customPaymentOutboxModel) MarkDone(ctx context.Context, id int64) error {
	query := fmt.Sprintf(`UPDATE %s SET status = 'done', done_at = now(), updated_at = now() WHERE id = $1`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// MarkDead 重试超限弃置 + 告警由调用方承接。
func (m *customPaymentOutboxModel) MarkDead(ctx context.Context, id int64, lastErr string) error {
	query := fmt.Sprintf(`UPDATE %s SET status = 'dead', last_error = $2, updated_at = now() WHERE id = $1`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, lastErr)
	return err
}

// IncrAttempt 尝试计数自增并按指数退避推进 next_retry_at，返回自增后的次数。
func (m *customPaymentOutboxModel) IncrAttempt(ctx context.Context, id int64, backoffSeconds int) (int64, error) {
	query := fmt.Sprintf(`UPDATE %s SET attempts = attempts + 1,
		next_retry_at = now() + ($3 || ' seconds')::interval,
		updated_at = now()
		WHERE id = $1
		RETURNING attempts`, m.table)
	var attempts int64
	err := m.conn.QueryRowCtx(ctx, &attempts, query, id, fmt.Sprint(backoffSeconds))
	if err != nil {
		return 0, err
	}
	return attempts, nil
}
