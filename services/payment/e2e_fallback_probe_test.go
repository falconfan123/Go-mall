//go:build e2eprobe

package main

// 结算兜底三层端到端探针（任务 7.2，tag 守卫不进常规测试/白名单）。
// 前置：order/inventory/coupons/payment 服务（新二进制）+ mall 库 + dtm 容器；
// payment 以快 tick 配置启动（relay 1s / scanner 2s，见 /tmp/fbprobe/payment.yaml）。
//
// 探针①：outbox 全链（翻转→relay→EnsureSaga→done，订单 Paid/库存扣减）
// 探针②：scan1 孤儿单自动补结算（无 outbox，超 10min 的 PAID+Pending）
// 探针③：墓碑单递增重试 r2/r3 → 超限 exception（券分支确定性失败构造）
// 探针④：scan2 超龄单自动关单（40min 仍待支付）
// 探针⑤：Stripe 对账——无本地 Stripe 测试环境，行为由单测覆盖（金额/币种拦截、
//         自动重放、方向2 冻结），live 探针留待 Stripe 测试环境（人工）。

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	inventorypb "github.com/falconfan123/Go-mall/services/inventory/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	_ "github.com/lib/pq"
)

const (
	fbDSN    = "postgres://root:fht3825099@localhost:5432/mall?sslmode=disable"
	fbDtmDSN = "postgres://root:fht3825099@localhost:5432/dtm?sslmode=disable"
	fbUser   = 8888
	fbProID  = 991001
	fbProID2 = 991002
)

func fbDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", fbDSN)
	if err != nil || db.Ping() != nil {
		t.Skipf("postgres unreachable, skip: %v", err)
	}
	return db
}

func fbSeed(t *testing.T, db *sql.DB, orderID, couponID string, orderStatus, payStatus, paidAt int64) {
	fbSeedAt(t, db, orderID, couponID, orderStatus, payStatus, paidAt, 40)
}

func fbSeedAt(t *testing.T, db *sql.DB, orderID, couponID string, orderStatus, payStatus, paidAt int64, createdMinAgo int) {
	t.Helper()
	ddb, err := sql.Open("postgres", fbDtmDSN)
	if err != nil || ddb.Ping() != nil {
		t.Skipf("dtm store unreachable, skip: %v", err)
	}
	defer ddb.Close()
	cleans := []string{
		fmt.Sprintf(`DELETE FROM orders WHERE order_id='%s'`, orderID),
		fmt.Sprintf(`DELETE FROM payments WHERE order_id='%s'`, orderID),
		fmt.Sprintf(`DELETE FROM payment_outbox WHERE order_id='%s'`, orderID),
		fmt.Sprintf(`DELETE FROM settlement_exception WHERE order_id='%s'`, orderID),
		fmt.Sprintf(`DELETE FROM order_items WHERE order_id='%s'`, orderID),
		fmt.Sprintf(`DELETE FROM order_addresses WHERE order_id='%s'`, orderID),
		fmt.Sprintf(`DELETE FROM user_coupons WHERE user_id=%d AND coupon_id='%s'`, fbUser, couponID),
		fmt.Sprintf(`DELETE FROM coupon_usage WHERE order_id='%s'`, orderID),
		fmt.Sprintf(`DELETE FROM inventory WHERE product_id IN (%d,%d)`, fbProID, fbProID2),
		`DELETE FROM inventory_lock WHERE order_id='pre-fb'`,
		`DELETE FROM return_lock WHERE order_id='pre-fb'`,
	}
	for _, q := range cleans {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("clean %s: %v", orderID, err)
		}
	}
	if _, err := db.Exec(`DELETE FROM dtm_barrier.barrier WHERE gid LIKE 'settle:%' AND gid LIKE '%` + orderID + `%'`); err != nil {
		t.Fatalf("clean barrier %s: %v", orderID, err)
	}
	for _, q := range []string{
		`DELETE FROM trans_global WHERE gid LIKE 'settle:%' AND gid LIKE '%` + orderID + `%'`,
		`DELETE FROM trans_branch_op WHERE gid LIKE 'settle:%' AND gid LIKE '%` + orderID + `%'`,
	} {
		if _, err := ddb.Exec(q); err != nil {
			t.Fatalf("clean dtm %s: %v", orderID, err)
		}
	}
	if _, err := db.Exec(`INSERT INTO orders (order_id, pre_order_id, user_id, coupon_id, payment_method, transaction_id, paid_at, original_amount, discount_amount, payable_amount, paid_amount, order_status, payment_status, reason, expire_time, created_at)
		VALUES ($1, 'pre-fb', $2, $3, 1, '', 0, 10000, 900, 9100, 0, $4, $5, '', 0, now() - ($6::text || ' minutes')::interval)`,
		orderID, fbUser, couponID, orderStatus, payStatus, fmt.Sprint(createdMinAgo)); err != nil {
		t.Fatalf("seed order %s: %v", orderID, err)
	}
	if _, err := db.Exec(`INSERT INTO order_items (order_id, product_id, quantity, price, product_name, product_desc)
		VALUES ($1, $2, 1, 100, 'probe-a', 'probe-desc-a'), ($1, $3, 2, 50, 'probe-b', 'probe-desc-b')`,
		orderID, fbProID, fbProID2); err != nil {
		t.Fatalf("seed items: %v", err)
	}
	// processStripePaymentSuccess 事务外预取 GetOrder 强校验订单地址快照，缺失即失败
	// （address_id 显式取 max+1：BIGSERIAL 序列落后于存量数据，默认 nextval 撞 pkey）
	if _, err := db.Exec(`INSERT INTO order_addresses (order_id, address_id, recipient_name, city, detailed_address)
		SELECT $1, COALESCE((SELECT max(address_id) + 1 FROM order_addresses), 1), 'probe', 'probe-city', 'probe-addr'
		ON CONFLICT (order_id) DO NOTHING`, orderID); err != nil {
		t.Fatalf("seed order address: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (product_id, total, sold) VALUES ($1, 100, 0), ($2, 100, 0)`,
		fbProID, fbProID2); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO payments (payment_id, pre_order_id, order_id, user_id, original_amount, expire_time, status, paid_at, paid_amount, transaction_id, payment_method)
		VALUES ($1, 'pre-fb', $2, $3, 10000, 0, 2, $4, 9100, 'tx-fb', 'stripe')`,
		"pay-"+orderID, orderID, fbUser, paidAt); err != nil {
		t.Fatalf("seed payment %s: %v", orderID, err)
	}
}

func fbWait(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func fbOrderStatus(t *testing.T, db *sql.DB, orderID string) int64 {
	t.Helper()
	var st int64
	if err := db.QueryRow(`SELECT order_status FROM orders WHERE order_id=$1`, orderID).Scan(&st); err != nil {
		return -1
	}
	return st
}

func fbSold(t *testing.T, db *sql.DB, pid int64) int64 {
	t.Helper()
	var sold int64
	_ = db.QueryRow(`SELECT sold FROM inventory WHERE product_id=$1`, pid).Scan(&sold)
	return sold
}

// 探针①：outbox 全链（seed pending outbox → relay 接力 → saga 完成 → done + 订单 Paid）
func TestE2EFallbackOutboxRelayChain(t *testing.T) {
	db := fbDB(t)
	fbSeed(t, db, "fb-o1", "", 2, 2, time.Now().Unix())
	input := &svc.SettlementInput{
		OrderID: "fb-o1", UserID: fbUser, PreOrderID: "pre-fb",
		TransactionID: "tx-fb-o1", PaidAmount: 9100, PaidAt: time.Now().Unix(),
		CouponID: "", OriginAmount: 10000, DiscountAmount: 900,
	}
	input.Items = append(input.Items,
		&inventorypb.InventoryReq_Items{ProductId: fbProID, Quantity: 1},
		&inventorypb.InventoryReq_Items{ProductId: fbProID2, Quantity: 2})
	payload, _ := json.Marshal(input)
	if _, err := db.Exec(`INSERT INTO payment_outbox (order_id, payment_id, user_id, payload, status)
		VALUES ('fb-o1', 'pay-fb-o1', $1, $2, 'pending')`, fbUser, string(payload)); err != nil {
		t.Fatalf("seed outbox: %v", err)
	}

	fbWait(t, 30*time.Second, func() bool {
		var st string
		if err := db.QueryRow(`SELECT status FROM payment_outbox WHERE order_id='fb-o1'`).Scan(&st); err != nil {
			return false
		}
		return st == "done"
	})
	fbWait(t, 30*time.Second, func() bool { return fbOrderStatus(t, db, "fb-o1") == 3 })
	if got := fbSold(t, db, fbProID); got != 1 {
		t.Fatalf("inventory 991001 sold = %d, want 1", got)
	}
	if got := fbSold(t, db, fbProID2); got != 2 {
		t.Fatalf("inventory 991002 sold = %d, want 2", got)
	}
}

// 探针②：scan1 孤儿单自动补结算（无 outbox 行，PAID 超 10min + 订单 Pending）
func TestE2EFallbackScan1ResettlesOrphan(t *testing.T) {
	db := fbDB(t)
	fbSeed(t, db, "fb-o2", "", 2, 2, time.Now().Add(-20*time.Minute).Unix())

	fbWait(t, 30*time.Second, func() bool { return fbOrderStatus(t, db, "fb-o2") == 3 })
	if got := fbSold(t, db, fbProID); got != 1 {
		t.Fatalf("inventory sold = %d, want 1", got)
	}
}

// 探针③：墓碑单递增重试 r2/r3 → 超限 exception（券分支确定性失败：无 user_coupons 行）
func TestE2EFallbackTombstoneEscalatesToException(t *testing.T) {
	db := fbDB(t)
	// paid_at = now-15min（过 scan1 阈值）；coupon_id 设置但 user_coupons 缺失 → 券分支必然 Aborted
	fbSeed(t, db, "fb-o3", "cpn-fb3", 2, 2, time.Now().Add(-15*time.Minute).Unix())

	// 等待 scan1 完成 tombstone r2/r3 重试并超限 → exception
	//（三次 saga 全程 ~90s：dtm 分支执行 + 补偿 + scanner tick）
	fbWait(t, 180*time.Second, func() bool {
		var cnt int
		if err := db.QueryRow(`SELECT count(*) FROM settlement_exception
			WHERE order_id='fb-o3' AND status='open'`).Scan(&cnt); err != nil {
			return false
		}
		return cnt == 1
	})
	var reason string
	if err := db.QueryRow(`SELECT reason FROM settlement_exception WHERE order_id='fb-o3' AND status='open'`).
		Scan(&reason); err != nil {
		t.Fatalf("query exception: %v", err)
	}
	if reason != "墓碑重试超限（r2..r{n} 均 failed）" {
		t.Fatalf("exception reason = %q", reason)
	}
	// 订单保持 PendingPayment（补偿已回滚）
	if st := fbOrderStatus(t, db, "fb-o3"); st != 2 {
		t.Fatalf("order should be PendingPayment after compensation, got %d", st)
	}
}

// 探针④：scan2 超龄单自动关单（40min 仍待支付，created_at 已按 40min 前播种）
func TestE2EFallbackScan2ClosesExpiredOrder(t *testing.T) {
	db := fbDB(t)
	// 无 payment（未支付），订单 PendingPayment + created_at 40min 前（seed 固定 -40min）
	fbSeed(t, db, "fb-o4", "", 2, 2, 0)
	if _, err := db.Exec(`DELETE FROM payments WHERE order_id='fb-o4'`); err != nil {
		t.Fatalf("remove payment: %v", err)
	}

	fbWait(t, 60*time.Second, func() bool { return fbOrderStatus(t, db, "fb-o4") == 6 })
}
