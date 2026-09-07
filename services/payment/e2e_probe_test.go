//go:build e2eprobe

package main

// 结算 Saga 端到端探针（任务 6.2，tag 守卫不进常规测试/白名单）。
// 前置：本地 dtm 容器（36790）+ order/inventory/coupons 服务（10004/10007/10009，
// 需为本 change 新二进制）+ postgres(mall) + rabbitmq。
// 运行：cd services/payment && go test -tags e2eprobe -run TestE2ESettlement -v .
// 探针①：正常结算三分支全成功（订单 Paid/库存扣减/券使用 + 流水）
// 探针②③：券分支确定性失败（用户券不存在→Aborted）→ 逆序补偿（库存回补/订单回滚）
// 探针④：终态 saga 重复提交 → dtm 报 cannot sumbmit（Ensure 映射由单测覆盖）
// 探针⑤：delay 队列毒消息一次即弃（由 rabbitmqadmin 单独执行，见 review-log）

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	couponspb "github.com/falconfan123/Go-mall/services/coupons/pb"
	inventorypb "github.com/falconfan123/Go-mall/services/inventory/pb"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	probeDSN   = "postgres://root:fht3825099@localhost:5432/mall?sslmode=disable"
	probeDTMDB = "postgres://root:fht3825099@localhost:5432/dtm?sslmode=disable"
	probeDTM   = "localhost:36790"
	orderURL   = "host.docker.internal:10004/order.OrderService/"
	inventURL  = "host.docker.internal:10007/inventory.Inventory/"
	couponsURL = "host.docker.internal:10009/coupons.Coupons/"
	probeUser  = 7777
)

func probeDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", probeDSN)
	if err != nil || db.Ping() != nil {
		t.Skipf("postgres unreachable, skip e2e probe: %v", err)
	}
	return db
}

func probeDtmDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", probeDTMDB)
	if err != nil || db.Ping() != nil {
		t.Skipf("dtm store unreachable, skip e2e probe: %v", err)
	}
	return db
}

func seedProbeData(t *testing.T, db *sql.DB, orderID, couponID string, withCoupon bool) {
	t.Helper()
	for _, q := range []string{
		fmt.Sprintf(`DELETE FROM orders WHERE order_id='%s'`, orderID),
		fmt.Sprintf(`DELETE FROM user_coupons WHERE user_id=%d AND coupon_id='%s'`, probeUser, couponID),
		fmt.Sprintf(`DELETE FROM coupon_usage WHERE order_id='%s'`, orderID),
		`DELETE FROM inventory_lock WHERE order_id='pre-e2e'`,
		`DELETE FROM return_lock WHERE order_id='pre-e2e'`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("clean probe data: %v", err)
		}
	}
	if _, err := db.Exec(`INSERT INTO orders (order_id, pre_order_id, user_id, coupon_id, payment_method, transaction_id, paid_at, original_amount, discount_amount, payable_amount, paid_amount, order_status, payment_status, reason, expire_time)
VALUES ($1, 'pre-e2e', $2, $3, 1, '', 0, 10000, 900, 9100, 0, 2, 2, '', 0)
ON CONFLICT (order_id) DO UPDATE SET coupon_id=$3, order_status=2, payment_status=2`, orderID, probeUser, couponID); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	for _, pid := range []int64{990001, 990002} {
		if _, err := db.Exec(`INSERT INTO inventory (product_id, total, sold) VALUES ($1, 100, 0)
ON CONFLICT (product_id) DO UPDATE SET total=100, sold=0`, pid); err != nil {
			t.Fatalf("seed inventory: %v", err)
		}
	}
	if _, err := db.Exec(`INSERT INTO coupons (id, name, type, value, min_amount, start_time, end_time, status, total_count, remaining_count) VALUES ($1, 'e2e-probe', 1, 900, 0, now() - interval '1 day', now() + interval '1 day', 1, 100, 100)
ON CONFLICT (id) DO UPDATE SET value=900, status=1`, couponID); err != nil {
		t.Fatalf("seed coupon: %v", err)
	}
	if withCoupon {
		if _, err := db.Exec(`INSERT INTO user_coupons (user_id, coupon_id, status, order_id) VALUES ($1, $2, 2, $3)
ON CONFLICT (order_id) DO UPDATE SET status=2`, probeUser, couponID, orderID); err != nil {
			t.Fatalf("seed user coupon: %v", err)
		}
	}
}

func submitSettlement(t *testing.T, orderID, couponID string) error {
	t.Helper()
	sg := dtmgrpc.NewSagaGrpc(probeDTM, "settle:"+orderID)
	sg.Add(orderURL+"UpdateOrder2PaymentSuccess", orderURL+"UpdateOrder2PaymentSuccessRollback",
		&orderpb.UpdateOrder2PaymentSuccessRequest{
			OrderId: orderID, UserId: probeUser,
			PaymentResult: &orderpb.PaymentResult{TransactionId: "tx-e2e", PaidAmount: 9100, PaidAt: time.Now().Unix()},
		})
	sg.Add(inventURL+"DecreaseInventory", inventURL+"ReturnInventory",
		&inventorypb.InventoryReq{
			Items:      []*inventorypb.InventoryReq_Items{{ProductId: 990001, Quantity: 1}, {ProductId: 990002, Quantity: 2}},
			PreOrderId: "pre-e2e", UserId: probeUser,
		})
	if couponID != "" {
		sg.Add(couponsURL+"UseCoupon", couponsURL+"UseCouponRollback",
			&couponspb.UseCouponReq{UserId: probeUser, PreOrderId: "pre-e2e", OrderId: orderID,
				CouponId: couponID, DiscountAmount: 900, OriginAmount: 10000})
	}
	return sg.Submit()
}

func sagaStatus(t *testing.T, ddb *sql.DB, orderID string) string {
	t.Helper()
	var st string
	if err := ddb.QueryRow(`SELECT status FROM trans_global WHERE gid=$1`, "settle:"+orderID).Scan(&st); err != nil {
		return ""
	}
	return st
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("condition not met before timeout")
}

func cleanDtmState(t *testing.T, db, ddb *sql.DB, gid string) {
	t.Helper()
	// barrier 表在业务库 mall；trans 表在 dtm 库
	if _, err := db.Exec(`DELETE FROM dtm_barrier.barrier WHERE gid=$1`, gid); err != nil {
		t.Fatalf("clean barrier: %v", err)
	}
	for _, q := range []string{
		`DELETE FROM trans_branch_op WHERE gid=$1`,
		`DELETE FROM trans_global WHERE gid=$1`,
	} {
		if _, err := ddb.Exec(q, gid); err != nil {
			t.Fatalf("clean dtm state: %v", err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM dtm_barrier.barrier WHERE gid=$1`, gid)
		_, _ = ddb.Exec(`DELETE FROM trans_branch_op WHERE gid=$1`, gid)
		_, _ = ddb.Exec(`DELETE FROM trans_global WHERE gid=$1`, gid)
	})
}

func orderStatus(t *testing.T, db *sql.DB, orderID string) (int64, int64) {
	t.Helper()
	var st, ps int64
	if err := db.QueryRow(`SELECT order_status, payment_status FROM orders WHERE order_id=$1`, orderID).Scan(&st, &ps); err != nil {
		t.Fatalf("query order: %v", err)
	}
	return st, ps
}

func soldOf(t *testing.T, db *sql.DB, pid int64) int64 {
	t.Helper()
	var sold int64
	if err := db.QueryRow(`SELECT sold FROM inventory WHERE product_id=$1`, pid).Scan(&sold); err != nil {
		t.Fatalf("query inventory: %v", err)
	}
	return sold
}

// 探针① 正常结算：三分支全成功 + 探针④ 终态重复提交报 cannot sumbmit
func TestE2ESettlementAllBranchesSucceed(t *testing.T) {
	db := probeDB(t)
	ddb := probeDtmDB(t)
	seedProbeData(t, db, "e2e-a", "cpn-e2e-a", true)
	cleanDtmState(t, db, ddb, "settle:e2e-a")

	if err := submitSettlement(t, "e2e-a", "cpn-e2e-a"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	waitFor(t, 90*time.Second, func() bool {
		st, _ := orderStatus(t, db, "e2e-a")
		return sagaStatus(t, ddb, "e2e-a") == "succeed" && st == 3
	})
	if st, ps := orderStatus(t, db, "e2e-a"); st != 3 || ps != 3 {
		t.Fatalf("order should be Paid/Paid, got %d/%d", st, ps)
	}
	if got := soldOf(t, db, 990001); got != 1 {
		t.Fatalf("inventory 990001 sold = %d, want 1", got)
	}
	if got := soldOf(t, db, 990002); got != 2 {
		t.Fatalf("inventory 990002 sold = %d, want 2", got)
	}
	var ucStatus int
	if err := db.QueryRow(`SELECT status FROM user_coupons WHERE user_id=$1 AND coupon_id='cpn-e2e-a'`, probeUser).Scan(&ucStatus); err != nil || ucStatus != 3 {
		t.Fatalf("coupon should be Used(3), got %d err=%v", ucStatus, err)
	}
	var usage int
	if err := db.QueryRow(`SELECT count(*) FROM coupon_usage WHERE order_id='e2e-a'`).Scan(&usage); err != nil || usage != 1 {
		t.Fatalf("usage record want 1, got %d err=%v", usage, err)
	}

	// 探针④：终态重复提交（模拟 webhook 快路径的 Ensure 原始行为）
	err := submitSettlement(t, "e2e-a", "cpn-e2e-a")
	if err == nil || status.Code(err) != codes.Aborted {
		t.Fatalf("finished saga resubmit should be rejected with Aborted, got %v", err)
	}
	fmt.Fprintln(os.Stderr, "probe④ dup-submit error:", err)
}

// 探针②③ 券分支确定性失败 → 逆序补偿：库存回补、订单回滚、无使用流水
func TestE2ESettlementCompensationOnCouponFailure(t *testing.T) {
	db := probeDB(t)
	ddb := probeDtmDB(t)
	// 不播种 user_coupons → UseCoupon 返回 CouponsNotExist（业务规则失败 → Aborted）
	seedProbeData(t, db, "e2e-b", "cpn-e2e-b", false)
	cleanDtmState(t, db, ddb, "settle:e2e-b")

	if err := submitSettlement(t, "e2e-b", "cpn-e2e-b"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	waitFor(t, 90*time.Second, func() bool { return sagaStatus(t, ddb, "e2e-b") == "failed" })
	if st, _ := orderStatus(t, db, "e2e-b"); st != 2 {
		t.Fatalf("order should be rolled back to PendingPayment(2), got %d", st)
	}
	if got := soldOf(t, db, 990001); got != 0 {
		t.Fatalf("inventory 990001 should be restored, got sold=%d", got)
	}
	if got := soldOf(t, db, 990002); got != 0 {
		t.Fatalf("inventory 990002 should be restored, got sold=%d", got)
	}
	var usage int
	if err := db.QueryRow(`SELECT count(*) FROM coupon_usage WHERE order_id='e2e-b'`).Scan(&usage); err != nil || usage != 0 {
		t.Fatalf("no usage record expected, got %d", usage)
	}
}
