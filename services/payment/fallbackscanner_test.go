package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// fakeRawConn：只读连接 fake（按查询片段分发结果）。
type fakeRawConn struct {
	sqlx.SqlConn
	scan1Rows []scan1Row
	scan2Rows []scan2Row
	payStatus int64 // recon: 按 payment_id 查状态
	payable   int64 // recon: 订单应付
	payableEr error
}

func (f *fakeRawConn) QueryRowsCtx(ctx context.Context, dest any, q string, args ...any) error {
	switch {
	case strings.Contains(q, "payments p JOIN orders"):
		if p, ok := dest.(*[]scan1Row); ok {
			*p = f.scan1Rows
		}
	case strings.Contains(q, "created_at < now()"):
		if p, ok := dest.(*[]scan2Row); ok {
			*p = f.scan2Rows
		}
	}
	return nil
}

func (f *fakeRawConn) QueryRowCtx(ctx context.Context, dest any, q string, args ...any) error {
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
	if f.payableEr != nil {
		return f.payableEr
	}
	return nil
}

// fakeDtm 假 dtm query API：gid 前缀 → 状态（缺省 ""）。
func fakeDtmServer(t *testing.T, statuses map[string]string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gid, _ := url.QueryUnescape(strings.TrimPrefix(r.URL.RawQuery, "gid="))
		st := statuses[gid]
		_ = json.NewEncoder(w).Encode(map[string]any{
			"transaction": map[string]string{"status": st},
		})
	}))
	t.Cleanup(ts.Close)
	return ts
}

func scannerFixture(t *testing.T, raw *fakeRawConn, saga *fakeSettlementSaga, statuses map[string]string,
	closeErr error) (*PaymentService, *fallbackScanner, *fakeExceptionModel) {
	t.Helper()
	ts := fakeDtmServer(t, statuses)
	exc := &fakeExceptionModel{}
	sc := &svc.ServiceContext{
		RawConn:        raw,
		ExceptionModel: exc,
		SettlementSaga: saga,
		OrderRpc: &fakeOrderRpc{
			getStateResp: pendingOrderDetail(), getDetailResp: orderDetailWithItems(),
			closeErr: closeErr,
		},
		Config: config.Config{
			Dtm:      config.DtmConfig{Server: "localhost:36790", HttpAddr: ts.URL},
			Fallback: config.FallbackConfig{Scan1ThresholdMinutes: 10, Scan2ThresholdMinutes: 35, TombstoneMaxRetry: 3}.Effective(),
		},
	}
	ps := &PaymentService{ctx: sc}
	return ps, newFallbackScanner(sc, sc.Config.Fallback.Effective()), exc
}

func scan1PendingRow(orderID string) scan1Row {
	return scan1Row{OrderID: orderID, UserID: 7, TransactionID: "tx-1", PaidAmount: 9900,
		PaidAt: 123, OrderStatus: 2, PreOrderID: "pre-1", CouponID: "", OriginalAmount: 10000}
}

// 无 saga 记录 → EnsureSaga（原始 gid，无后缀）
func TestScan1OrphanEnsuresSaga(t *testing.T) {
	initLog(t)
	raw := &fakeRawConn{scan1Rows: []scan1Row{scan1PendingRow("o-s1")}}
	saga := &fakeSettlementSaga{}
	_, sc, exc := scannerFixture(t, raw, saga, map[string]string{"settle:o-s1": ""}, nil)

	sc.scan1(context.Background())

	if saga.ensureN != 1 || saga.initiateN != 0 {
		t.Fatalf("orphan must Ensure once, got ensure=%d initiate=%d", saga.ensureN, saga.initiateN)
	}
	if saga.inputs[0].GidSuffix != "" {
		t.Fatalf("orphan ensure must use original gid, got %q", saga.inputs[0].GidSuffix)
	}
	if len(exc.upserts) != 0 {
		t.Fatalf("no exception expected: %+v", exc.upserts)
	}
}

// in-flight → 干等不干预
func TestScan1InFlightWaits(t *testing.T) {
	initLog(t)
	raw := &fakeRawConn{scan1Rows: []scan1Row{scan1PendingRow("o-s2")}}
	saga := &fakeSettlementSaga{}
	_, sc, exc := scannerFixture(t, raw, saga, map[string]string{"settle:o-s2": "submitted"}, nil)

	sc.scan1(context.Background())

	if saga.ensureN != 0 || saga.initiateN != 0 {
		t.Fatalf("in-flight must not intervene: ensure=%d initiate=%d", saga.ensureN, saga.initiateN)
	}
	if len(exc.upserts) != 0 {
		t.Fatalf("in-flight must not alert: %+v", exc.upserts)
	}
}

// 墓碑 failed → r2 未用 → Initiate(GidSuffix=":r2")
func TestScan1TombstoneRetriesWithSuffix(t *testing.T) {
	initLog(t)
	raw := &fakeRawConn{scan1Rows: []scan1Row{scan1PendingRow("o-s3")}}
	saga := &fakeSettlementSaga{}
	_, sc, _ := scannerFixture(t, raw, saga, map[string]string{
		"settle:o-s3": "failed", "settle:o-s3:r2": "",
	}, nil)

	sc.scan1(context.Background())

	if saga.initiateN != 1 || saga.ensureN != 0 {
		t.Fatalf("tombstone must Initiate with suffix, got initiate=%d ensure=%d", saga.initiateN, saga.ensureN)
	}
	if saga.inputs[0].GidSuffix != ":r2" {
		t.Fatalf("suffix = %q, want :r2", saga.inputs[0].GidSuffix)
	}
}

// 墓碑全部 failed（r2/r3）→ 超限 exception 转人工
func TestScan1TombstoneOverLimit(t *testing.T) {
	initLog(t)
	raw := &fakeRawConn{scan1Rows: []scan1Row{scan1PendingRow("o-s4")}}
	saga := &fakeSettlementSaga{}
	_, sc, exc := scannerFixture(t, raw, saga, map[string]string{
		"settle:o-s4": "failed", "settle:o-s4:r2": "failed", "settle:o-s4:r3": "failed",
	}, nil)

	sc.scan1(context.Background())

	if saga.initiateN != 0 {
		t.Fatalf("over-limit must not initiate, got %d", saga.initiateN)
	}
	if len(exc.upserts) != 1 || !strings.Contains(exc.upserts[0].reason, "墓碑重试超限") {
		t.Fatalf("over-limit must alert: %+v", exc.upserts)
	}
}

// succeed 但订单非 Paid → 数据矛盾 exception
func TestScan1SucceedContradiction(t *testing.T) {
	initLog(t)
	raw := &fakeRawConn{scan1Rows: []scan1Row{scan1PendingRow("o-s5")}}
	saga := &fakeSettlementSaga{}
	_, sc, exc := scannerFixture(t, raw, saga, map[string]string{"settle:o-s5": "succeed"}, nil)

	sc.scan1(context.Background())

	if saga.ensureN != 0 || saga.initiateN != 0 {
		t.Fatalf("contradiction must not re-drive settlement")
	}
	if len(exc.upserts) != 1 || !strings.Contains(exc.upserts[0].reason, "数据矛盾") {
		t.Fatalf("must record contradiction: %+v", exc.upserts)
	}
}

// 关闭订单的已支付款 → 退款人工产出（迟到支付竞态的收敛终点）
func TestScan1ClosedOrderProducesRefundTask(t *testing.T) {
	initLog(t)
	row := scan1PendingRow("o-s6")
	row.OrderStatus = int64(6) // Closed
	raw := &fakeRawConn{scan1Rows: []scan1Row{row}}
	saga := &fakeSettlementSaga{}
	_, sc, exc := scannerFixture(t, raw, saga, nil, nil)

	sc.scan1(context.Background())

	if saga.ensureN != 0 || saga.initiateN != 0 {
		t.Fatalf("closed order must not resettle")
	}
	if len(exc.upserts) != 1 || !strings.Contains(exc.upserts[0].reason, "需退款") {
		t.Fatalf("must produce refund task: %+v", exc.upserts)
	}
}

// scan2：超龄单调用关单 RPC；连续失败 3 轮产出 exception
func TestScan2ClosesAndEscalatesAfterConsecutiveFailures(t *testing.T) {
	initLog(t)
	raw := &fakeRawConn{scan2Rows: []scan2Row{{OrderID: "o-s7", UserID: 9}}}
	saga := &fakeSettlementSaga{}
	_, sc, exc := scannerFixture(t, raw, saga, nil, errors.New("order rpc down"))

	sc.scan2(context.Background()) // fail 1
	sc.scan2(context.Background()) // fail 2
	if len(exc.upserts) != 0 {
		t.Fatalf("below threshold must not alert: %+v", exc.upserts)
	}
	sc.scan2(context.Background()) // fail 3 → exception
	if len(exc.upserts) != 1 || exc.upserts[0].source != "scan2" {
		t.Fatalf("3rd consecutive failure must escalate: %+v", exc.upserts)
	}

	// 成功路径：换无错误 fixture
	raw2 := &fakeRawConn{scan2Rows: []scan2Row{{OrderID: "o-s8", UserID: 9}}}
	_, sc2, exc2 := scannerFixture(t, raw2, saga, nil, nil)
	sc2.scan2(context.Background())
	if len(exc2.upserts) != 0 {
		t.Fatalf("success must not alert")
	}
}
