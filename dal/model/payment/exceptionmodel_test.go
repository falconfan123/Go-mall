package payment

import (
	"context"
	"database/sql"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// TestExceptionUpsertAggregation 跨层聚合去重（C8）：同单 upsert 计数+来源去重追加；
// resolved 后同单新异常开新行（处置后不重复告警、复发再告）。
func TestExceptionUpsertAggregation(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://root:fht3825099@localhost:5432/mall?sslmode=disable")
	if err != nil {
		t.Skipf("postgres open failed, skip: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("postgres unreachable, skip: %v", err)
	}
	defer db.Close()

	m := NewSettlementExceptionModel(sqlx.NewSqlConn("postgres", "postgres://root:fht3825099@localhost:5432/mall?sslmode=disable"))
	orderID := "exc-test-order"
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM settlement_exception WHERE order_id=$1`, orderID) })
	_, _ = db.Exec(`DELETE FROM settlement_exception WHERE order_id=$1`, orderID)
	ctx := context.Background()

	// 首次发现：relay
	seen, isNew, err := m.UpsertOpen(ctx, orderID, "p-1", "relay", "outbox retry limit exceeded", nil)
	if err != nil || !isNew || seen != 1 {
		t.Fatalf("first upsert = seen=%d isNew=%v err=%v", seen, isNew, err)
	}
	// 重复发现（scan1）：计数递增、不新增行、sources 去重追加
	seen, isNew, err = m.UpsertOpen(ctx, orderID, "p-1", "scan1", "outbox retry limit exceeded", nil)
	if err != nil || isNew || seen != 2 {
		t.Fatalf("dup upsert = seen=%d isNew=%v err=%v", seen, isNew, err)
	}
	seen, isNew, _ = m.UpsertOpen(ctx, orderID, "p-1", "relay", "outbox retry limit exceeded", nil)
	if isNew || seen != 2 {
		t.Fatalf("duplicate source must not double-append: seen=%d isNew=%v", seen, isNew)
	}
	exc, err := m.FindOpenByOrderId(ctx, orderID)
	if err != nil {
		t.Fatalf("find open: %v", err)
	}
	if exc.SeenCount != 2 {
		t.Fatalf("seen_count = %d, want 2", exc.SeenCount)
	}

	// 处置：resolved
	if err := m.MarkResolved(ctx, orderID, "人工核实完成"); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if _, err := m.FindOpenByOrderId(ctx, orderID); err == nil {
		t.Fatal("resolved order should have no open exception")
	}

	// 复发：开新行
	seen, isNew, err = m.UpsertOpen(ctx, orderID, "p-1", "scan2", "关单连续失败", nil)
	if err != nil || !isNew || seen != 1 {
		t.Fatalf("recurrence must open new row: seen=%d isNew=%v err=%v", seen, isNew, err)
	}
}
