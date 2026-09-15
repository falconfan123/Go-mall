package logic

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	ordertypes "github.com/falconfan123/Go-mall/common/types/order"
	ordermodel "github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/order/internal/svc"
	order "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ---------- fakes（直连路径单元测试） ----------

type fakeOrdersModel struct {
	ordermodel.OrdersModel
	order   *ordermodel.Orders
	orderEr error
	updated int
	updateF error
}

func (f *fakeOrdersModel) WithSession(sqlx.Session) ordermodel.OrdersModel { return f }

func (f *fakeOrdersModel) GetOrderByOrderIDAndUserIDWithLock(context.Context, string, int32) (*ordermodel.Orders, error) {
	return f.order, f.orderEr
}

func (f *fakeOrdersModel) UpdateOrder2Payment(context.Context, string, int32,
	*ordertypes.PaymentResult, ordertypes.OrderStatus, ordertypes.PaymentStatus) error {
	f.updated++
	return f.updateF
}

func newFakeLogic(t *testing.T, om *fakeOrdersModel) *UpdateOrder2PaymentSuccessLogic {
	t.Helper()
	// 直连路径的事务骨架用真实 conn（fake model 拦截全部业务查询，事务内无真实语句）
	conn := sqlx.NewSqlConn("postgres", testDSN)
	return &UpdateOrder2PaymentSuccessLogic{
		ctx: context.Background(),
		svcCtx: &svc.ServiceContext{
			OrderModel: om,
			Model:      conn,
			DtmDB:      openRealDB(t),
		},
		Logger: testLogger(t),
	}
}

func validReq(orderID string) *order.UpdateOrder2PaymentSuccessRequest {
	return &order.UpdateOrder2PaymentSuccessRequest{
		OrderId: orderID,
		UserId:  7,
		PaymentResult: &order.PaymentResult{
			TransactionId: "tx-1",
			PaidAmount:    100,
			PaidAt:        time.Now().Unix(),
		},
	}
}

func pendingOrder() *ordermodel.Orders {
	return &ordermodel.Orders{
		OrderId:       "o-1",
		UserId:        7,
		OrderStatus:   int64(ordertypes.OrderStatusPendingPayment),
		PaymentStatus: int64(ordertypes.PaymentStatusPaying),
	}
}

// 待支付订单正常流转 Paid
func TestSettlePendingPaymentTransitions(t *testing.T) {
	om := &fakeOrdersModel{order: pendingOrder()}
	l := newFakeLogic(t, om)
	res, err := l.UpdateOrder2PaymentSuccess(validReq("o-1"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.StatusCode != code.Success {
		t.Fatalf("status = %d, want Success", res.StatusCode)
	}
	if om.updated != 1 {
		t.Fatalf("order should be updated once, got %d", om.updated)
	}
}

// 已 Paid 重放必须返回成功（不触发二次更新、不 Aborted——spec"已支付重放不触发补偿"）
func TestSettlePaidReplayReturnsSuccess(t *testing.T) {
	o := pendingOrder()
	o.OrderStatus = int64(ordertypes.OrderStatusPaid)
	o.PaymentStatus = int64(ordertypes.PaymentStatusPaid)
	om := &fakeOrdersModel{order: o}
	l := newFakeLogic(t, om)
	res, err := l.UpdateOrder2PaymentSuccess(validReq("o-1"))
	if err != nil {
		t.Fatalf("paid replay must NOT be an error (would mis-trigger dtm compensation): %v", err)
	}
	if res.StatusCode != code.Success {
		t.Fatalf("paid replay status = %d, want Success", res.StatusCode)
	}
	if om.updated != 0 {
		t.Fatalf("paid replay must not update again, got %d updates", om.updated)
	}
}

// 状态冲突（Closed）：确定性失败 → codes.Aborted（触发 dtm 补偿）
func TestSettleConflictStatusAborted(t *testing.T) {
	o := pendingOrder()
	o.OrderStatus = int64(ordertypes.OrderStatusClosed)
	o.PaymentStatus = int64(ordertypes.PaymentStatusExpired)
	om := &fakeOrdersModel{order: o}
	l := newFakeLogic(t, om)
	_, err := l.UpdateOrder2PaymentSuccess(validReq("o-1"))
	if err == nil || status.Code(err) != codes.Aborted {
		t.Fatalf("conflict must return Aborted, got %v", err)
	}
	if om.updated != 0 {
		t.Fatalf("conflict must not update order")
	}
}

// 订单不存在：确定性失败 → codes.Aborted（design 决策 2 悬挂风险声明）
func TestSettleOrderNotFoundAborted(t *testing.T) {
	om := &fakeOrdersModel{order: nil, orderEr: sqlx.ErrNotFound}
	l := newFakeLogic(t, om)
	_, err := l.UpdateOrder2PaymentSuccess(validReq("o-404"))
	if err == nil || status.Code(err) != codes.Aborted {
		t.Fatalf("not-found must return Aborted, got %v", err)
	}
}

// 基础设施失败：返回 Internal（dtm 退避重试，不触发补偿）
func TestSettleInfraErrorInternal(t *testing.T) {
	om := &fakeOrdersModel{order: pendingOrder(), orderEr: fmt.Errorf("db down")}
	l := newFakeLogic(t, om)
	_, err := l.UpdateOrder2PaymentSuccess(validReq("o-1"))
	if err == nil || status.Code(err) != codes.Internal {
		t.Fatalf("infra error must return Internal, got %v", err)
	}
}

// ---------- barrier 路径（真实 postgres 集成测试） ----------

const testDSN = "postgres://root:fht3825099@localhost:5432/mall?sslmode=disable"

func newBarrierTestCtx(gid, op string) context.Context {
	md := metadata.Pairs(
		"dtm-gid", gid,
		"dtm-trans_type", "saga",
		"dtm-branch_id", "01",
		"dtm-op", op,
	)
	return metadata.NewIncomingContext(context.Background(), md)
}

func openRealDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Skipf("open postgres failed, skip: %v", err)
	}
	if err = db.Ping(); err != nil {
		t.Skipf("postgres unreachable, skip: %v", err)
	}
	return db
}

func barrierLogic(t *testing.T, db *sql.DB, ctx context.Context) *UpdateOrder2PaymentSuccessLogic {
	t.Helper()
	conn := sqlx.NewSqlConn("postgres", testDSN)
	return &UpdateOrder2PaymentSuccessLogic{
		ctx: ctx,
		svcCtx: &svc.ServiceContext{
			OrderModel: ordermodel.NewOrdersModel(conn),
			Model:      conn,
			DtmDB:      db,
		},
		Logger: testLogger(t),
	}
}

func seedOrder(t *testing.T, db *sql.DB, orderID string, status ordertypes.OrderStatus, payStatus ordertypes.PaymentStatus) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO orders (order_id, pre_order_id, user_id, coupon_id, payment_method, transaction_id, paid_at, original_amount, discount_amount, payable_amount, paid_amount, order_status, payment_status, reason, expire_time)
VALUES ($1, 'pre-test', 7, '', 1, '', 0, 100, 0, 100, 0, $2, $3, '', 0)
ON CONFLICT (order_id) DO UPDATE SET order_status = $2, payment_status = $3, paid_amount = 0, transaction_id = ''`,
		orderID, int(status), int(payStatus))
	if err != nil {
		t.Fatalf("seed order: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM orders WHERE order_id = $1`, orderID) })
}

func orderState(t *testing.T, db *sql.DB, orderID string) (int64, int64) {
	t.Helper()
	var st, ps int64
	if err := db.QueryRow(`SELECT order_status, payment_status FROM orders WHERE order_id = $1`, orderID).
		Scan(&st, &ps); err != nil {
		t.Fatalf("query order: %v", err)
	}
	return st, ps
}

func cleanBarrier(t *testing.T, db *sql.DB, gid string) {
	t.Helper()
	if _, err := db.Exec(`DELETE FROM dtm_barrier.barrier WHERE gid = $1`, gid); err != nil {
		t.Fatalf("clean barrier: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM dtm_barrier.barrier WHERE gid = $1`, gid) })
}

func barrierCount(t *testing.T, db *sql.DB, gid string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM dtm_barrier.barrier WHERE gid = $1`, gid).Scan(&n); err != nil {
		t.Fatalf("count barrier: %v", err)
	}
	return n
}

// 首次结算：PendingPayment → Paid，barrier 记录写入
func TestBarrierPathSettlesFirstTime(t *testing.T) {
	db := openRealDB(t)
	cleanBarrier(t, db, "gid-b1")
	seedOrder(t, db, "o-barrier-1", ordertypes.OrderStatusPendingPayment, ordertypes.PaymentStatusPaying)
	l := barrierLogic(t, db, newBarrierTestCtx("gid-b1", "action"))

	res, err := l.UpdateOrder2PaymentSuccess(validReq("o-barrier-1"))
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if res.StatusCode != code.Success {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if st, ps := orderState(t, db, "o-barrier-1"); st != int64(ordertypes.OrderStatusPaid) {
		t.Fatalf("order not paid: %d/%d", st, ps)
	}
	if barrierCount(t, db, "gid-b1") == 0 {
		t.Fatalf("barrier row must be recorded")
	}
}

// 重放：barrier 短路 → 返回成功、订单不再变更、不误触发补偿
func TestBarrierPathReplayShortCircuits(t *testing.T) {
	db := openRealDB(t)
	cleanBarrier(t, db, "gid-b2")
	seedOrder(t, db, "o-barrier-2", ordertypes.OrderStatusPendingPayment, ordertypes.PaymentStatusPaying)
	l := barrierLogic(t, db, newBarrierTestCtx("gid-b2", "action"))

	if _, err := l.UpdateOrder2PaymentSuccess(validReq("o-barrier-2")); err != nil {
		t.Fatalf("first settle: %v", err)
	}
	st1, _ := orderState(t, db, "o-barrier-2")

	// 模拟 saga 重放：同 gid 新 barrier_id 再次调用
	l2 := barrierLogic(t, db, newBarrierTestCtx("gid-b2", "action"))
	res, err := l2.UpdateOrder2PaymentSuccess(validReq("o-barrier-2"))
	if err != nil {
		t.Fatalf("replay must succeed (not Aborted): %v", err)
	}
	if res.StatusCode != code.Success {
		t.Fatalf("replay status = %d", res.StatusCode)
	}
	st2, _ := orderState(t, db, "o-barrier-2")
	if st1 != st2 {
		t.Fatalf("replay must not change order state: %d vs %d", st1, st2)
	}
}

// 冲突订单（Closed）：确定性失败 → Aborted（补偿），订单不变
func TestBarrierPathConflictAborts(t *testing.T) {
	db := openRealDB(t)
	cleanBarrier(t, db, "gid-b3")
	seedOrder(t, db, "o-barrier-3", ordertypes.OrderStatusClosed, ordertypes.PaymentStatusExpired)
	l := barrierLogic(t, db, newBarrierTestCtx("gid-b3", "action"))

	_, err := l.UpdateOrder2PaymentSuccess(validReq("o-barrier-3"))
	if err == nil || status.Code(err) != codes.Aborted {
		t.Fatalf("conflict must be Aborted, got %v", err)
	}
}

// 空补偿（null compensate）：op=compensate 无先行 action → barrier 直接返回成功，业务不执行
func TestBarrierPathNullCompensateNoOp(t *testing.T) {
	db := openRealDB(t)
	cleanBarrier(t, db, "gid-b4")
	seedOrder(t, db, "o-barrier-4", ordertypes.OrderStatusPendingPayment, ordertypes.PaymentStatusPaying)
	l := barrierLogic(t, db, newBarrierTestCtx("gid-b4", "compensate"))

	res, err := l.UpdateOrder2PaymentSuccess(validReq("o-barrier-4"))
	if err != nil {
		t.Fatalf("null compensate must succeed: %v", err)
	}
	if res.StatusCode != code.Success {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if st, _ := orderState(t, db, "o-barrier-4"); st != int64(ordertypes.OrderStatusPendingPayment) {
		t.Fatalf("null compensate must not touch order")
	}
}

func testLogger(t *testing.T) logx.Logger {
	t.Helper()
	return logx.WithContext(context.Background())
}
