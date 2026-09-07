package consumerx

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/utils/idempotency"
	"github.com/streadway/amqp"
	"github.com/zeromicro/go-zero/core/logx"
)

// ---------- fakes ----------

type fakeStore struct {
	mu        sync.Mutex
	states    map[string]idempotency.ClaimState
	marked    []string // MarkDone 成功的 key
	released  []string
	counters  map[string]int64
	claimErr  error
	markErr   error
	releaseN  int // Release 可注入失败次数（>0 时前 N 次 Release 返回错误）
	releasedN int
}

func newFakeStore() *fakeStore {
	return &fakeStore{states: map[string]idempotency.ClaimState{}, counters: map[string]int64{}}
}

func (f *fakeStore) Claim(_ context.Context, key string, _ time.Duration) (idempotency.ClaimState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.claimErr != nil {
		return idempotency.ClaimNew, f.claimErr
	}
	st, ok := f.states[key]
	if !ok {
		st = idempotency.ClaimNew
		f.states[key] = st
	}
	return st, nil
}

func (f *fakeStore) MarkDone(_ context.Context, key string, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.markErr != nil {
		return f.markErr
	}
	f.states[key] = idempotency.ClaimDone
	f.marked = append(f.marked, key)
	return nil
}

func (f *fakeStore) Release(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.releasedN++
	if f.releaseN > 0 && f.releasedN <= f.releaseN {
		return errors.New("release failed (injected)")
	}
	delete(f.states, key)
	f.released = append(f.released, key)
	return nil
}

func (f *fakeStore) IncrWithTTL(_ context.Context, key string, _ time.Duration) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.counters[key]++
	return f.counters[key], nil
}

type fakeAck struct {
	mu       sync.Mutex
	acked    int
	requeues int // Reject(true)
	drops    int // Reject(false)
}

func (f *fakeAck) Ack(bool) error { f.mu.Lock(); f.acked++; f.mu.Unlock(); return nil }

func (f *fakeAck) Reject(requeue bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if requeue {
		f.requeues++
	} else {
		f.drops++
	}
	return nil
}

var errBoom = errors.New("boom")

type testMsg struct {
	OrderID string
	Panic   bool
}

// fakeHandler：消息体 "ok:<orderId>[-panic]"；前 failN 次成功路径之外的 Handle 返回 errBoom。
type fakeHandler struct {
	mu    sync.Mutex
	callN int
	failN int
}

func (h *fakeHandler) Decode(body []byte) (testMsg, error) {
	s := string(body)
	if !strings.HasPrefix(s, "ok:") {
		return testMsg{}, errors.New("bad json")
	}
	id := strings.TrimPrefix(s, "ok:")
	m := testMsg{OrderID: id}
	if strings.HasSuffix(id, "-panic") {
		m.Panic = true
		m.OrderID = strings.TrimSuffix(id, "-panic")
	}
	return m, nil
}

func (h *fakeHandler) IdempotencyKey(m testMsg) string { return m.OrderID }

func (h *fakeHandler) Handle(_ context.Context, m testMsg) error {
	h.mu.Lock()
	h.callN++
	n := h.callN
	failN := h.failN
	h.mu.Unlock()

	if m.Panic {
		panic("handler panic")
	}
	if failN > 0 && n <= failN {
		return errBoom
	}
	return nil
}

// ---------- unit tests ----------

func newTestConsumer(t *testing.T, store Store, h *fakeHandler) (*Consumer[testMsg], *fakeAck) {
	t.Helper()
	logx.SetUp(logx.LogConf{Stat: false})
	ack := &fakeAck{}
	cfg := Config{
		QueueName:          "order-notify-queue",
		IdempotencyService: "order",
		IdempotencyQueue:   "notify",
		RetryLimit:         3,
	}
	return New[testMsg](nil, nil, store, h, cfg), ack
}

// 成功 -> MarkDone(SUCCESS) + Ack，且不释放 claim
func TestProcessSuccessAcks(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("ok:o1"))

	if ack.acked != 1 || ack.requeues != 0 || ack.drops != 0 {
		t.Fatalf("want acked=1, got %+v", ack)
	}
	if len(store.marked) != 1 || store.marked[0] != "idempotency:order:notify:o1" {
		t.Fatalf("success path must MarkDone, got %v", store.marked)
	}
	if len(store.released) != 0 {
		t.Fatalf("success path must not release claim, got %v", store.released)
	}
	if store.states["idempotency:order:notify:o1"] != idempotency.ClaimDone {
		t.Fatalf("claim state should be Done, got %v", store.states["idempotency:order:notify:o1"])
	}
}

// 毒消息 -> 一次即弃，不触碰重试计数与幂等键
func TestProcessPoisonDropsOnce(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("not-a-valid-payload"))

	if ack.drops != 1 || ack.requeues != 0 || ack.acked != 0 {
		t.Fatalf("want drop=1, got %+v", ack)
	}
	if len(store.counters) != 0 || len(store.states) != 0 {
		t.Fatalf("poison path must not touch retry counter or claim: %v %v", store.counters, store.states)
	}
}

// 业务失败未超限 -> 计数 + Release 幂等 key + Reject(true)
func TestProcessFailureReleasesAndRequeues(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{failN: 1}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("ok:o2"))

	if ack.requeues != 1 || ack.acked != 0 || ack.drops != 0 {
		t.Fatalf("want requeue=1, got %+v", ack)
	}
	key := "idempotency:order:notify:o2" // BuildKey 输出含前缀（既有约定，audit 双前缀兼容，C13）
	if len(store.released) != 1 || store.released[0] != key {
		t.Fatalf("claim must be released before requeue, got %v", store.released)
	}
	if store.counters["settlement:retry:notify:o2"] != 1 {
		t.Fatalf("retry counter want 1, got %v", store.counters)
	}
}

// 达到重试上限 -> 弃单 Reject(false)，不释放 claim
func TestProcessFailureAbandonsAtLimit(t *testing.T) {
	store := newFakeStore()
	store.counters["settlement:retry:notify:o3"] = 2 // 本次是第 3 次失败，达到上限 3
	h := &fakeHandler{failN: 99}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("ok:o3"))

	if ack.drops != 1 || ack.requeues != 0 {
		t.Fatalf("want drop=1, got %+v", ack)
	}
	if len(store.released) != 0 {
		t.Fatalf("abandoned message must not release claim: %v", store.released)
	}
}

// 失败路径日志级别：重投=可恢复级（info），弃单=需人工介入级（error），可区分
func TestFailureLogLevelsDistinguishable(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{failN: 1}
	c, ack := newTestConsumer(t, store, h)

	var mu sync.Mutex
	var buf strings.Builder
	logx.SetWriter(logx.NewWriter(writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return buf.Write(p)
	})))
	defer func() { logx.SetUp(logx.LogConf{Stat: false}) }()

	c.Process(context.Background(), ack, []byte("ok:o9"))

	mu.Lock()
	logs := buf.String()
	mu.Unlock()
	if !strings.Contains(logs, "message requeued for retry") || !strings.Contains(strings.ToLower(logs), "info") {
		t.Fatalf("requeue path must log at info level, got: %s", logs)
	}

	// 第 3 次失败达到上限 -> error 级弃单日志
	store2 := newFakeStore()
	store2.counters["settlement:retry:notify:o9"] = 2
	h2 := &fakeHandler{failN: 99}
	c2, ack2 := newTestConsumer(t, store2, h2)
	mu.Lock()
	buf.Reset()
	mu.Unlock()
	c2.Process(context.Background(), ack2, []byte("ok:o9"))
	mu.Lock()
	logs = buf.String()
	mu.Unlock()
	if !strings.Contains(logs, "message abandoned after retry limit") || !strings.Contains(strings.ToLower(logs), "error") {
		t.Fatalf("abandon path must log at error level, got: %s", logs)
	}
}

// claim 失败（结果未知）：不执行业务，按可重试失败重投
func TestProcessClaimErrorRequeues(t *testing.T) {
	store := newFakeStore()
	store.claimErr = errors.New("redis down")
	h := &fakeHandler{}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("ok:o4"))

	if ack.requeues != 1 {
		t.Fatalf("want requeue=1, got %+v", ack)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.callN != 0 {
		t.Fatalf("business handler must not run when claim fails")
	}
}

// MarkDone 失败：按业务失败处理（重投后业务幂等收敛），消息不 Ack
func TestProcessMarkDoneFailureRequeues(t *testing.T) {
	store := newFakeStore()
	store.markErr = errors.New("redis down on mark")
	h := &fakeHandler{}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("ok:o7"))

	if ack.requeues != 1 || ack.acked != 0 {
		t.Fatalf("mark-done failure should requeue, got %+v", ack)
	}
	if len(store.released) != 1 {
		t.Fatalf("claim should be released for redelivery: %v", store.released)
	}
}

// panic 不外泄，按业务失败处理（重投）
func TestProcessPanicRecovered(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("ok:o5-panic"))

	if ack.requeues != 1 || ack.drops != 0 {
		t.Fatalf("panic should be treated as retryable failure, got %+v", ack)
	}
}

// 已成功消息（Done）重投 -> Ack 跳过，不重复执行业务
func TestProcessDuplicateSkips(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("ok:o6"))
	c.Process(context.Background(), ack, []byte("ok:o6"))

	if ack.acked != 2 {
		t.Fatalf("both deliveries should ack: %+v", ack)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.callN != 1 {
		t.Fatalf("duplicate should not re-handle, got handled=%d", h.callN)
	}
}

// 在途消息（Processing，租约未过期）重投 -> Ack 跳过
func TestProcessProcessingStateSkips(t *testing.T) {
	store := newFakeStore()
	store.states["idempotency:order:notify:o8"] = idempotency.ClaimProcessing
	h := &fakeHandler{}
	c, ack := newTestConsumer(t, store, h)

	c.Process(context.Background(), ack, []byte("ok:o8"))

	if ack.acked != 1 || ack.requeues != 0 {
		t.Fatalf("processing-state message should ack-skip, got %+v", ack)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.callN != 0 {
		t.Fatalf("processing state must not re-handle business")
	}
}

// ---------- 优雅退出与监督 ----------

func TestRunStopsOnContextCancel(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{}
	c, _ := newTestConsumer(t, store, h)
	c.cfg.ShutdownTimeout = 200 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { c.Run(ctx); close(done) }()

	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}

// 无连接且无 Factory：consumeOnce 立即失败，Run 按退避重启（恢复路径）
func TestRunRestartsWithBackoffWithoutConn(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{}
	c, _ := newTestConsumer(t, store, h)
	c.cfg.BackoffBase = 10 * time.Millisecond
	c.cfg.ShutdownTimeout = 200 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { c.Run(ctx); close(done) }()

	time.Sleep(150 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}

// ensureConn：无连接 + Factory 时走重拨（此处断言错误透传）
func TestEnsureConnRedialsViaFactory(t *testing.T) {
	store := newFakeStore()
	h := &fakeHandler{}
	c, _ := newTestConsumer(t, store, h)
	c.conn = nil
	c.factory = func() (*amqp.Connection, error) {
		return nil, errors.New("dial failed (expected in unit test)")
	}
	if _, err := c.ensureConn(); err == nil {
		t.Fatal("expected dial error")
	}
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
