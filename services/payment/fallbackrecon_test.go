package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// recon fakeRawConn：本地集合/订单应付/支付单状态按查询分发。
type reconRawConn struct {
	sqlx.SqlConn
	localRows []reconLocalPaid
	payStatus int64
	payable   int64
}

func (f *reconRawConn) QueryRowsCtx(ctx context.Context, dest any, q string, args ...any) error {
	if p, ok := dest.(*[]reconLocalPaid); ok && strings.Contains(q, "paid_at >= $1") {
		*p = f.localRows
	}
	return nil
}

func (f *reconRawConn) QueryRowCtx(ctx context.Context, dest any, q string, args ...any) error {
	switch {
	case strings.Contains(q, "SELECT status FROM payments"):
		if p, ok := dest.(*int64); ok {
			*p = f.payStatus
		}
	case strings.Contains(q, "payable_amount"):
		if p, ok := dest.(*int64); ok {
			*p = f.payable
		}
	}
	return nil
}

func reconFixture(t *testing.T, raw *reconRawConn, saga *fakeSettlementSaga, sessions []stripeSessionData) (*PaymentService, *fallbackScanner, *fakeExceptionModel) {
	t.Helper()
	exc := &fakeExceptionModel{}
	pm := &fakePaymentModel{found: unpaidPaymentRecord(), condRows: 1}
	sc := &svc.ServiceContext{
		RawConn:        raw,
		ExceptionModel: exc,
		SettlementSaga: saga,
		PaymentModel:   pm,
		OutboxModel:    newFakeOutboxModel(),
		Model:          &fakeTxRunner{execRows: 1},
		OrderRpc: &fakeOrderRpc{
			getStateResp:  pendingOrderDetail(),
			getDetailResp: orderDetailWithItems(),
		},
		Config: config.Config{Fallback: config.FallbackConfig{ReconHour: 2}.Effective()},
	}
	ps := &PaymentService{ctx: sc}
	// 注入 fake Stripe 拉取（Rule 6）
	ps.stripeLister = func(ctx context.Context, startSec, endSec int64) ([]stripeSessionData, error) {
		return sessions, nil
	}
	return ps, newFallbackScanner(sc, sc.Config.Fallback.Effective()), exc
}

func runRecon(t *testing.T, ps *PaymentService) {
	t.Helper()
	now := time.Now()
	// 直接驱动当日逻辑：把 ReconHour 调成当前小时
	ps.ctx.Config.Fallback.ReconHour = now.Hour()
	ps.reconOnce(context.Background(), ps.ctx.Config.Fallback)
	if ps.lastReconDay == "" {
		t.Fatal("recon did not run")
	}
}

// 方向1：Stripe 有账、本地未 PAID、金额一致 → 自动重放（Initiate，订单待支付）
func TestReconDirection1ReplaysLostPayment(t *testing.T) {
	initLog(t)
	raw := &reconRawConn{payStatus: 1 /*未支付*/, payable: 9900}
	saga := &fakeSettlementSaga{}
	sessions := []stripeSessionData{{
		TxKey: "tx-stripe-1", AmountTotal: 9900, Currency: "cny",
		OrderID: "o-r1", PaymentID: "p-r1", UserID: 7, CreatedAtUnix: 111,
	}}
	ps, _, exc := reconFixture(t, raw, saga, sessions)
	// 修复：ProcessStripePaymentSuccess 的本地集合查询走 RawConn（局部 PAID 列表），
	// 方向1 依赖 transaction_id 查询——fake 未实现该分支则视为缺失 ✓
	runRecon(t, ps)

	if saga.initiateN != 1 || saga.ensureN != 0 {
		t.Fatalf("lost payment must be replayed via Initiate, got initiate=%d ensure=%d", saga.initiateN, saga.ensureN)
	}
	if saga.inputs[0].OrderID != "o-r1" || saga.inputs[0].TransactionID != "tx-stripe-1" {
		t.Fatalf("replay input mismatch: %+v", saga.inputs[0])
	}
	if len(exc.upserts) != 0 {
		t.Fatalf("successful replay must not alert: %+v", exc.upserts)
	}
}

// 方向1 金额不一致 → 停止重放转人工
func TestReconDirection1AmountMismatchStops(t *testing.T) {
	initLog(t)
	raw := &reconRawConn{payStatus: 1, payable: 100 /*与 Stripe 9900 不一致*/}
	saga := &fakeSettlementSaga{}
	sessions := []stripeSessionData{{
		TxKey: "tx-stripe-2", AmountTotal: 9900, Currency: "cny",
		OrderID: "o-r2", PaymentID: "p-r2", UserID: 7,
	}}
	ps, _, exc := reconFixture(t, raw, saga, sessions)
	runRecon(t, ps)

	if saga.initiateN != 0 {
		t.Fatalf("amount mismatch must stop replay, got %d", saga.initiateN)
	}
	if len(exc.upserts) != 1 || !strings.Contains(exc.upserts[0].reason, "金额与本地应付不一致") {
		t.Fatalf("must alert on amount mismatch: %+v", exc.upserts)
	}
}

// 方向1 币种不一致 → 停止重放转人工
func TestReconDirection1CurrencyMismatchStops(t *testing.T) {
	initLog(t)
	raw := &reconRawConn{payStatus: 1, payable: 9900}
	saga := &fakeSettlementSaga{}
	sessions := []stripeSessionData{{
		TxKey: "tx-stripe-3", AmountTotal: 9900, Currency: "usd",
		OrderID: "o-r3", PaymentID: "p-r3", UserID: 7,
	}}
	ps, _, exc := reconFixture(t, raw, saga, sessions)
	runRecon(t, ps)

	if saga.initiateN != 0 {
		t.Fatalf("currency mismatch must stop replay")
	}
	if len(exc.upserts) != 1 || !strings.Contains(exc.upserts[0].reason, "币种不一致") {
		t.Fatalf("must alert on currency mismatch: %+v", exc.upserts)
	}
}

// 方向2：本地 PAID、Stripe 无账 → 告警冻结，不触发任何结算动作
func TestReconDirection2Freezes(t *testing.T) {
	initLog(t)
	raw := &reconRawConn{localRows: []reconLocalPaid{{
		PaymentID: "p-r4", OrderID: "o-r4", UserID: 7, TransactionID: "tx-local-4",
	}}}
	saga := &fakeSettlementSaga{}
	ps, sc, exc := reconFixture(t, raw, saga, nil) // Stripe 无账

	sc.scan1(context.Background()) // 不应触发
	// recon：
	now := time.Now()
	ps.ctx.Config.Fallback.ReconHour = now.Hour()
	ps.reconOnce(context.Background(), ps.ctx.Config.Fallback)

	if saga.initiateN != 0 || saga.ensureN != 0 {
		t.Fatalf("direction2 must not drive settlement")
	}
	if len(exc.upserts) != 1 || !strings.Contains(exc.upserts[0].reason, "Stripe 无对应账") {
		t.Fatalf("must freeze with alert: %+v", exc.upserts)
	}
}

// 当日幂等：同日二次 reconOnce 跳过
func TestReconRunsOncePerDay(t *testing.T) {
	initLog(t)
	raw := &reconRawConn{}
	saga := &fakeSettlementSaga{}
	ps, sc, _ := reconFixture(t, raw, saga, nil)
	runRecon(t, ps)
	first := ps.lastReconDay
	ps.reconOnce(context.Background(), ps.ctx.Config.Fallback)
	if ps.lastReconDay != first {
		t.Fatalf("same-day recon must be skipped")
	}
	_ = sc
}
