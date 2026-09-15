package inventory

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const invTestDSN = "postgres://root:fht3825099@localhost:5432/mall?sslmode=disable"

// TestCompensationReleasesDecreaseLock 7.2 探针③ live 缺陷回归：
// saga 失败补偿（ReturnInventory）必须释放扣减锁——否则墓碑重试的新 gid
// （settle:{order}:r2）扣减阶段1 命中陈旧锁幂等跳过，saga 报 succeed 但库存未扣。
// 验证：扣减 → 回补 → 扣减锁已删 → 同单可再次扣减。
func TestCompensationReleasesDecreaseLock(t *testing.T) {
	db, err := sql.Open("postgres", invTestDSN)
	if err != nil {
		t.Skipf("postgres open failed, skip: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("postgres unreachable, skip: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConn("postgres", invTestDSN)
	m := NewInventoryModel(conn)
	const (
		pid    = int32(88990001)
		order  = "inv-lock-test-order"
		userID = int64(7)
	)
	db.Exec(`DELETE FROM inventory WHERE product_id=$1`, pid)
	db.Exec(`DELETE FROM inventory_lock WHERE order_id=$1`, order)
	db.Exec(`DELETE FROM return_lock WHERE order_id=$1`, order)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM inventory WHERE product_id=$1`, pid)
		db.Exec(`DELETE FROM inventory_lock WHERE order_id=$1`, order)
		db.Exec(`DELETE FROM return_lock WHERE order_id=$1`, order)
	})
	if _, err := db.Exec(`INSERT INTO inventory (product_id, total, sold) VALUES ($1, 100, 0)`, pid); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	pids := []int32{pid}
	qtys := []int32{2}
	ctx := context.Background()

	// 扣减（建扣减锁）
	if err := m.BatchDecreaseInventoryAtom(ctx, pids, qtys, userID, order); err != nil {
		t.Fatalf("decrease: %v", err)
	}
	var sold int64
	if err := db.QueryRow(`SELECT sold FROM inventory WHERE product_id=$1`, pid).Scan(&sold); err != nil || sold != 2 {
		t.Fatalf("sold after decrease = %d err=%v, want 2", sold, err)
	}

	// 补偿（应释放扣减锁 + 恢复库存）
	if err := m.BatchReturnInventoryAtom(ctx, pids, qtys, order, userID); err != nil {
		t.Fatalf("return: %v", err)
	}
	if err := db.QueryRow(`SELECT sold FROM inventory WHERE product_id=$1`, pid).Scan(&sold); err != nil || sold != 0 {
		t.Fatalf("sold after return = %d err=%v, want 0", sold, err)
	}
	var lockCnt int
	if err := db.QueryRow(`SELECT count(*) FROM inventory_lock WHERE order_id=$1 AND user_id=$2`, order, userID).
		Scan(&lockCnt); err != nil || lockCnt != 0 {
		t.Fatalf("decrease lock after compensation = %d err=%v, want 0（补偿未释放扣减锁）", lockCnt, err)
	}

	// 同单重试（墓碑 r2 语义）可重新扣减
	if err := m.BatchDecreaseInventoryAtom(ctx, pids, qtys, userID, order); err != nil {
		t.Fatalf("re-decrease after compensation: %v", err)
	}
	if err := db.QueryRow(`SELECT sold FROM inventory WHERE product_id=$1`, pid).Scan(&sold); err != nil || sold != 2 {
		t.Fatalf("sold after re-decrease = %d err=%v, want 2", sold, err)
	}
}
