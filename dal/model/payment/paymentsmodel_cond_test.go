package payment

import (
	"context"
	"database/sql"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const condTestDSN = "postgres://root:fht3825099@localhost:5432/mall?sslmode=disable"

// TestUpdateStatusToPaidWithSession 条件更新门槛（真实库，skip 模式同 idempotency 测试）：
// 首次翻转 rows=1；对已 PAID 重复翻转 rows=0（并发 webhook 只有一次生效）。
func TestUpdateStatusToPaidWithSession(t *testing.T) {
	db, err := sql.Open("postgres", condTestDSN)
	if err != nil {
		t.Skipf("postgres open failed, skip: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("postgres unreachable, skip: %v", err)
	}
	defer db.Close()

	m := NewPaymentsModel(sqlx.NewSqlConn("postgres", condTestDSN))
	pid := "cond-test-payment"
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM payments WHERE payment_id=$1`, pid) })
	_, _ = db.Exec(`DELETE FROM payments WHERE payment_id=$1`, pid)
	if _, err := db.Exec(`INSERT INTO payments (payment_id, pre_order_id, user_id, original_amount, expire_time, status)
		VALUES ($1,'pre-cond',7,1000,0,1)`, pid); err != nil {
		t.Fatalf("seed: %v", err)
	}

	conn := sqlx.NewSqlConn("postgres", condTestDSN)
	var got int64
	if err := conn.TransactCtx(context.Background(), func(ctx context.Context, session sqlx.Session) error {
		var txErr error
		got, txErr = m.UpdateStatusToPaidWithSession(ctx, session, pid, "tx-1", 990, 123)
		return txErr
	}); err != nil {
		t.Fatalf("first update: %v", err)
	}
	if got != 1 {
		t.Fatalf("first flip rows = %d, want 1", got)
	}
	var status int64
	if err := db.QueryRow(`SELECT status FROM payments WHERE payment_id=$1`, pid).Scan(&status); err != nil || status != 2 {
		t.Fatalf("status = %d err=%v, want 2", status, err)
	}
	var txID sql.NullString
	if err := db.QueryRow(`SELECT transaction_id FROM payments WHERE payment_id=$1`, pid).Scan(&txID); err != nil || txID.String != "tx-1" {
		t.Fatalf("transaction_id = %v err=%v", txID, err)
	}

	if err := conn.TransactCtx(context.Background(), func(ctx context.Context, session sqlx.Session) error {
		var txErr error
		got, txErr = m.UpdateStatusToPaidWithSession(ctx, session, pid, "tx-2", 990, 456)
		return txErr
	}); err != nil {
		t.Fatalf("second update: %v", err)
	}
	if got != 0 {
		t.Fatalf("second flip rows = %d, want 0 (状态机门槛)", got)
	}
}
