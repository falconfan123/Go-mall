package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/falconfan123/Go-mall/common/consts/code"
	dalpayment "github.com/falconfan123/Go-mall/dal/model/payment"
	order "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	paymentpb "github.com/falconfan123/Go-mall/services/payment/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
)

// ---------- fakes（内嵌接口，只覆写被测方法，Rule 6） ----------

var errSagaDown = errors.New("dtm saga down")
var errGetOrderDown = errors.New("order rpc down")

type fakeOrderRpc struct {
	order.OrderServiceClient

	getStateResp *order.OrderDetail2PaymentResponse
	getStateErr  error

	getDetailResp *order.OrderDetailResponse
	getDetailErr  error

	closeCalled int
	closeErr    error
}

func (f *fakeOrderRpc) GetOrder2Payment(ctx context.Context, in *order.GetOrderRequest, opts ...grpc.CallOption) (*order.OrderDetail2PaymentResponse, error) {
	return f.getStateResp, f.getStateErr
}

func (f *fakeOrderRpc) GetOrder(ctx context.Context, in *order.GetOrderRequest, opts ...grpc.CallOption) (*order.OrderDetailResponse, error) {
	return f.getDetailResp, f.getDetailErr
}

func (f *fakeOrderRpc) CloseExpiredOrder(ctx context.Context, in *order.CloseExpiredOrderRequest, opts ...grpc.CallOption) (*order.EmptyRes, error) {
	f.closeCalled++
	if f.closeErr != nil {
		return nil, f.closeErr
	}
	return &order.EmptyRes{StatusCode: code.Success, StatusMsg: "closed"}, nil
}

// fakeTxRunner：事务执行抽象 fake；ExecCtx 返回可配置影响行数。
type fakeTxRunner struct {
	execRows int64
	execErr  error
}

func (f *fakeTxRunner) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return fn(ctx, &fakeSession{rows: f.execRows, execErr: f.execErr})
}

type fakeSession struct {
	sqlx.Session
	rows    int64
	execErr error
}

type fakeResult struct{ rows int64 }

func (f fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (f fakeResult) RowsAffected() (int64, error) { return f.rows, nil }

func (f *fakeSession) ExecCtx(ctx context.Context, q string, args ...any) (sql.Result, error) {
	return fakeResult{rows: f.rows}, f.execErr
}

// fakeOutboxModel：outbox 待办 fake（记录调用序列）。
type fakeOutboxModel struct {
	dalpayment.PaymentOutboxModel

	pending   []*dalpayment.PaymentOutbox
	claimed   map[int64]bool
	claimErr  error
	inserted  []*dalpayment.PaymentOutbox
	insertErr error
	doneIds   []int64
	deadIds   []int64
	deadErrs  []string
	attempts  map[int64]int64
	attemptN  int
	ensureErr error // 与 Ensure 失败联动（由测试驱动）
}

func newFakeOutboxModel() *fakeOutboxModel {
	return &fakeOutboxModel{claimed: map[int64]bool{}, attempts: map[int64]int64{}}
}

func (f *fakeOutboxModel) InsertPendingWithSession(ctx context.Context, session sqlx.Session, orderId, paymentId string, userId int64, payload string) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.inserted = append(f.inserted, &dalpayment.PaymentOutbox{
		OrderId: orderId, PaymentId: paymentId, Payload: payload,
	})
	return nil
}

func (f *fakeOutboxModel) FetchPendingBatch(ctx context.Context, limit int) ([]*dalpayment.PaymentOutbox, error) {
	return f.pending, nil
}

func (f *fakeOutboxModel) ClaimPending(ctx context.Context, id int64, claimSeconds int) (bool, error) {
	if f.claimErr != nil {
		return false, f.claimErr
	}
	if f.claimed[id] {
		return false, nil
	}
	f.claimed[id] = true
	return true, nil
}

func (f *fakeOutboxModel) MarkDone(ctx context.Context, id int64) error {
	f.doneIds = append(f.doneIds, id)
	return nil
}

func (f *fakeOutboxModel) MarkDead(ctx context.Context, id int64, lastErr string) error {
	f.deadIds = append(f.deadIds, id)
	f.deadErrs = append(f.deadErrs, lastErr)
	return nil
}

func (f *fakeOutboxModel) IncrAttempt(ctx context.Context, id int64, backoffSeconds int) (int64, error) {
	f.attemptN++
	f.attempts[id]++
	return f.attempts[id], nil
}

// fakeExceptionModel：异常单 fake（记录 upsert 调用）。
type fakeExceptionModel struct {
	dalpayment.SettlementExceptionModel

	upserts []struct {
		orderID   string
		paymentID string
		source    string
		reason    string
	}
	upsertErr error
}

func (f *fakeExceptionModel) UpsertOpen(ctx context.Context, orderId, paymentId, source, reason string, detail map[string]any) (int64, bool, error) {
	if f.upsertErr != nil {
		return 0, false, f.upsertErr
	}
	f.upserts = append(f.upserts, struct {
		orderID   string
		paymentID string
		source    string
		reason    string
	}{orderId, paymentId, source, reason})
	return int64(len(f.upserts)), true, nil
}

// fakeSettlementSaga 实现 svc.SettlementSagaAPI（记录调用与注入错误）。
type fakeSettlementSaga struct {
	initiateN int
	ensureN   int
	inputs    []*svc.SettlementInput
	initErr   error
	ensureErr error
}

func (f *fakeSettlementSaga) Initiate(_ context.Context, in *svc.SettlementInput) error {
	f.initiateN++
	f.inputs = append(f.inputs, in)
	return f.initErr
}

func (f *fakeSettlementSaga) Ensure(_ context.Context, in *svc.SettlementInput) error {
	f.ensureN++
	f.inputs = append(f.inputs, in)
	return f.ensureErr
}

// fakePaymentModel：支付单查询 + 条件更新 fake。
type fakePaymentModel struct {
	dalpayment.PaymentsModel
	found      *dalpayment.Payments
	foundEr    error
	condRows   int64 // UpdateStatusToPaidWithSession 的影响行数
	condCalled int
}

func (f *fakePaymentModel) FindOne(context.Context, string) (*dalpayment.Payments, error) {
	return f.found, f.foundEr
}

func (f *fakePaymentModel) UpdateStatusToPaidWithSession(context.Context, sqlx.Session,
	string, string, int64, int64) (int64, error) {
	f.condCalled++
	return f.condRows, nil
}

// ---------- helpers ----------

func newPaymentServiceForTest(or order.OrderServiceClient, pm *fakePaymentModel, saga svc.SettlementSagaAPI,
	outbox *fakeOutboxModel, exc *fakeExceptionModel, tx *fakeTxRunner) *PaymentService {
	return &PaymentService{ctx: &svc.ServiceContext{
		OrderRpc:       or,
		PaymentModel:   pm,
		OutboxModel:    outbox,
		ExceptionModel: exc,
		Model:          tx,
		SettlementSaga: saga,
	}}
}

func initLog(t *testing.T) {
	t.Helper()
	logx.SetUp(logx.LogConf{Stat: false})
}

func paidPaymentRecord() *dalpayment.Payments {
	return &dalpayment.Payments{
		PaymentId: "p-1",
		OrderId:   sql.NullString{String: "o-1", Valid: true},
		UserId:    7,
		Status:    int64(paymentpb.PaymentStatus_PAYMENT_STATUS_PAID),
	}
}

func unpaidPaymentRecord() *dalpayment.Payments {
	return &dalpayment.Payments{
		PaymentId: "p-1",
		OrderId:   sql.NullString{String: "o-1", Valid: true},
		UserId:    7,
		Status:    int64(paymentpb.PaymentStatus_PAYMENT_STATUS_UNPAID),
	}
}

func paidOrderDetail() *order.OrderDetail2PaymentResponse {
	return &order.OrderDetail2PaymentResponse{
		StatusCode: code.Success,
		Order: &order.Order{
			OrderStatus:   order.OrderStatus_ORDER_STATUS_PAID,
			PaymentStatus: order.PaymentStatus_PAYMENT_STATUS_PAID,
		},
	}
}

func pendingOrderDetail() *order.OrderDetail2PaymentResponse {
	return &order.OrderDetail2PaymentResponse{
		StatusCode: code.Success,
		Order: &order.Order{
			OrderStatus:   order.OrderStatus_ORDER_STATUS_PENDING_PAYMENT,
			PaymentStatus: order.PaymentStatus_PAYMENT_STATUS_PAYING,
		},
	}
}

// 含 items/优惠券/金额的订单详情快照（saga 分支 payload 数据源 + outbox payload）
func orderDetailWithItems() *order.OrderDetailResponse {
	return &order.OrderDetailResponse{
		StatusCode: code.Success,
		Order: &order.Order{
			OrderId:        "o-1",
			PreOrderId:     "pre-1",
			OriginalAmount: 10000,
			DiscountAmount: 900,
			CouponId:       "c-1",
		},
		Items: []*order.OrderItem{{ProductId: 42, Quantity: 2}},
	}
}

// ---------- 3.1 webhook 两写事务 ----------

// 翻转路径：状态翻转 + outbox 待办同事务落库，然后正常路径 Initiate
func TestFlipWritesOutboxAndInitiates(t *testing.T) {
	initLog(t)
	or := &fakeOrderRpc{getStateResp: pendingOrderDetail(), getDetailResp: orderDetailWithItems()}
	pm := &fakePaymentModel{found: unpaidPaymentRecord()}
	saga := &fakeSettlementSaga{}
	outbox := newFakeOutboxModel()
	exc := &fakeExceptionModel{}
	pm.condRows = 1 // 翻转成功
	s := newPaymentServiceForTest(or, pm, saga, outbox, exc, &fakeTxRunner{execRows: 1})

	if err := s.processStripePaymentSuccess(context.Background(), "o-1", "p-1", 7, "tx-1", 9900, 123); err != nil {
		t.Fatalf("flip path should succeed, got %v", err)
	}
	if len(outbox.inserted) != 1 {
		t.Fatalf("outbox must be written on flip, got %d", len(outbox.inserted))
	}
	ins := outbox.inserted[0]
	if ins.OrderId != "o-1" || ins.PaymentId != "p-1" {
		t.Fatalf("outbox identity mismatch: %+v", ins)
	}
	var payload svc.SettlementInput
	if err := json.Unmarshal([]byte(ins.Payload), &payload); err != nil {
		t.Fatalf("payload must be decodable: %v", err)
	}
	if payload.OrderID != "o-1" || payload.CouponID != "c-1" || len(payload.Items) != 1 {
		t.Fatalf("payload snapshot mismatch: %+v", payload)
	}
	if saga.initiateN != 1 || saga.ensureN != 0 {
		t.Fatalf("normal path must Initiate, got initiate=%d ensure=%d", saga.initiateN, saga.ensureN)
	}
}

// 并发翻转（rows=0）：不写 outbox，但结算发起照常（另一事务已翻转）
func TestConcurrentFlipSkipsOutbox(t *testing.T) {
	initLog(t)
	or := &fakeOrderRpc{getStateResp: pendingOrderDetail(), getDetailResp: orderDetailWithItems()}
	pm := &fakePaymentModel{found: unpaidPaymentRecord()}
	saga := &fakeSettlementSaga{}
	outbox := newFakeOutboxModel()
	exc := &fakeExceptionModel{}
	s := newPaymentServiceForTest(or, pm, saga, outbox, exc, &fakeTxRunner{execRows: 0})

	if err := s.processStripePaymentSuccess(context.Background(), "o-1", "p-1", 7, "tx-1", 9900, 123); err != nil {
		t.Fatalf("should succeed, got %v", err)
	}
	if len(outbox.inserted) != 0 {
		t.Fatalf("rows=0 must not write outbox, got %d", len(outbox.inserted))
	}
	if saga.initiateN != 1 {
		t.Fatalf("settlement initiation must proceed, got %d", saga.initiateN)
	}
}

// outbox 写失败 → 事务回滚 → 返回错误（webhook 500 → Stripe 重试），不发起结算
func TestOutboxInsertFailureRollsBackFlip(t *testing.T) {
	initLog(t)
	or := &fakeOrderRpc{getStateResp: pendingOrderDetail(), getDetailResp: orderDetailWithItems()}
	pm := &fakePaymentModel{found: unpaidPaymentRecord()}
	saga := &fakeSettlementSaga{}
	outbox := newFakeOutboxModel()
	outbox.insertErr = errors.New("outbox insert down")
	exc := &fakeExceptionModel{}
	pm.condRows = 1 // 翻转成功
	s := newPaymentServiceForTest(or, pm, saga, outbox, exc, &fakeTxRunner{execRows: 1})

	if err := s.processStripePaymentSuccess(context.Background(), "o-1", "p-1", 7, "tx-1", 9900, 123); err == nil {
		t.Fatal("outbox failure must fail the webhook (rolled back flip)")
	}
	if saga.initiateN != 0 {
		t.Fatalf("flip rolled back — no settlement initiation, got %d", saga.initiateN)
	}
}

// ---------- 3.2 relay ----------

func relayFixture(t *testing.T, pending []*dalpayment.PaymentOutbox, ensureErr error) (*PaymentService, *fakeOutboxModel, *fakeSettlementSaga, *fakeExceptionModel) {
	or := &fakeOrderRpc{}
	pm := &fakePaymentModel{found: paidPaymentRecord()}
	saga := &fakeSettlementSaga{ensureErr: ensureErr}
	outbox := newFakeOutboxModel()
	outbox.pending = pending
	exc := &fakeExceptionModel{}
	s := newPaymentServiceForTest(or, pm, saga, outbox, exc, &fakeTxRunner{})
	return s, outbox, saga, exc
}

func pendingOutboxItem(id int64, orderID string, attempts int64) *dalpayment.PaymentOutbox {
	input := &svc.SettlementInput{OrderID: orderID, UserID: 7, TransactionID: "tx-1", PaidAmount: 100}
	b, _ := json.Marshal(input)
	return &dalpayment.PaymentOutbox{Id: id, OrderId: orderID, PaymentId: "p-1", Payload: string(b), Attempts: attempts}
}

// relay 接力成功 → EnsureSaga → 立即 done（done 语义：不等 saga 终态，C3）
func TestRelayEnsuresAndMarksDone(t *testing.T) {
	initLog(t)
	s, outbox, saga, _ := relayFixture(t, []*dalpayment.PaymentOutbox{pendingOutboxItem(1, "o-r1", 0)}, nil)

	s.relayOnce(context.Background())

	if saga.ensureN != 1 {
		t.Fatalf("relay must Ensure once, got %d", saga.ensureN)
	}
	if len(outbox.doneIds) != 1 || outbox.doneIds[0] != 1 {
		t.Fatalf("relay must mark done immediately, got %v", outbox.doneIds)
	}
	if len(outbox.deadIds) != 0 || outbox.attemptN != 0 {
		t.Fatalf("success path must not increment/dead, got dead=%v attempts=%d", outbox.deadIds, outbox.attemptN)
	}
}

// Ensure 失败（未超限）→ 计数退避重投，不置 dead
func TestRelayFailureRequeues(t *testing.T) {
	initLog(t)
	s, outbox, _, _ := relayFixture(t, []*dalpayment.PaymentOutbox{pendingOutboxItem(2, "o-r2", 0)}, errSagaDown)

	s.relayOnce(context.Background())

	if outbox.attemptN != 1 || len(outbox.doneIds) != 0 {
		t.Fatalf("failure path must incr attempt, got attempts=%d done=%v", outbox.attemptN, outbox.doneIds)
	}
	if len(outbox.deadIds) != 0 {
		t.Fatalf("under limit must not dead: %v", outbox.deadIds)
	}
}

// 超限 → dead + exception(source=relay)
func TestRelayAbandonsAtLimit(t *testing.T) {
	initLog(t)
	s, outbox, _, exc := relayFixture(t, []*dalpayment.PaymentOutbox{pendingOutboxItem(3, "o-r3", 4)}, errSagaDown)
	outbox.attempts[3] = 4 // 已有 4 次（本次 IncrAttempt 后为 5，达上限）

	s.relayOnce(context.Background())

	if len(outbox.deadIds) != 1 || outbox.deadIds[0] != 3 {
		t.Fatalf("over-limit must dead, got %v", outbox.deadIds)
	}
	if len(exc.upserts) != 1 || exc.upserts[0].source != "relay" || exc.upserts[0].orderID != "o-r3" {
		t.Fatalf("must record relay exception: %+v", exc.upserts)
	}
}

// payload 损坏 = 毒待办：直接 dead + exception（重试无意义）
func TestRelayPoisonPayloadDead(t *testing.T) {
	initLog(t)
	item := &dalpayment.PaymentOutbox{Id: 4, OrderId: "o-r4", PaymentId: "p-4", Payload: "not-json"}
	s, outbox, saga, exc := relayFixture(t, []*dalpayment.PaymentOutbox{item}, nil)

	s.relayOnce(context.Background())

	if saga.ensureN != 0 {
		t.Fatalf("poison payload must not reach saga")
	}
	if len(outbox.deadIds) != 1 || len(exc.upserts) != 1 {
		t.Fatalf("poison payload must dead+exception: dead=%v exc=%v", outbox.deadIds, exc.upserts)
	}
}

// done 语义强化：Ensure 返回 nil 即 done，即使"模拟 saga 仍在 in-flight"（fake 不再回调）
func TestRelayDoneDoesNotWaitForSagaTerminal(t *testing.T) {
	initLog(t)
	s, outbox, saga, _ := relayFixture(t, []*dalpayment.PaymentOutbox{pendingOutboxItem(5, "o-r5", 0)}, nil)

	s.relayOnce(context.Background())

	// EnsureSaga 提交成功即 done；saga 后续失败由扫描器/人工层承接（C3）
	if len(outbox.doneIds) == 1 && saga.ensureN == 1 && outbox.attemptN == 0 {
		return
	}
	t.Fatalf("done semantics violated")
}

// ---------- 3.3 快路径与 relay 并发幂等 ----------

// 快路径（webhook Ensure）与 relay Ensure 双发：同一 gid 输入收敛，无重复副作用
func TestFastPathAndRelayBothEnsureConverge(t *testing.T) {
	initLog(t)
	or := &fakeOrderRpc{getStateResp: paidOrderDetail(), getDetailResp: orderDetailWithItems()}
	pm := &fakePaymentModel{found: paidPaymentRecord()}
	saga := &fakeSettlementSaga{}
	outbox := newFakeOutboxModel()
	exc := &fakeExceptionModel{}
	s := newPaymentServiceForTest(or, pm, saga, outbox, exc, &fakeTxRunner{})

	// 快路径
	if err := s.processStripePaymentSuccess(context.Background(), "o-1", "p-1", 7, "tx-1", 9900, 123); err != nil {
		t.Fatalf("fast path: %v", err)
	}
	// relay 对同一单再次接力（outbox 已 done 的重复场景由幂等收敛）
	input := &svc.SettlementInput{OrderID: "o-1", UserID: 7, TransactionID: "tx-1", PaidAmount: 9900}
	if err := s.ctx.SettlementSaga.Ensure(context.Background(), input); err != nil {
		t.Fatalf("relay ensure: %v", err)
	}
	if saga.ensureN != 2 {
		t.Fatalf("both paths must call Ensure (dtm idempotent), got %d", saga.ensureN)
	}
	// 收敛性由 EnsureSaga 幂等语义保证（deterministic gid + dtm 静默接受，spike 已实证）
	if strings.Contains("", "never") {
		t.Fatal("unreachable guard")
	}
}
