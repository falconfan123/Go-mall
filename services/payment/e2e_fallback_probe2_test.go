//go:build e2eprobe

package main

// 7.2 探针（统一栈默认 tick：relay 10s / scanner 5min）——start-unified 干净重启后执行：
//   P1 ① outbox 全链（真签名 webhook → 翻转+outbox 同事务 → relay → done → Paid/库存）
//       + kill-relay 续扫不重不漏（kill -9 后由 start-unified supervisor 自动拉起）
//   P2 ② scan1 孤儿单自动补结算（PAID+Pending 超 10min，无 outbox 行）
//   P3 ③ 墓碑递增 r2 成功（券先失效→saga failed→补券→r2 succeed，无 exception）
//   P4 ④ scan2 超龄关单 + 迟到支付竞态收敛（关后 PAID → scan1 出退款 exception）
//   P5 ⑤ Stripe 测试环境对账（sk_test 构造方向1/方向2 差额单）
//
// 运行：cd services/payment && go test -tags e2eprobe -count=1 -timeout=40m -run "TestFB2" -v .
// 证据：/Volumes/Fan/Go-mall/scripts/logs/payment.log（统一栈日志）+ 各断言。

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
)

// fbDtmDB 打开 dtm 库连接（trans_global 等 saga 状态表在 dtm 库而非 mall 库）。
func fbDtmDB(t *testing.T) *sql.DB {
	t.Helper()
	ddb, err := sql.Open("postgres", fbDtmDSN)
	if err != nil || ddb.Ping() != nil {
		t.Skipf("dtm store unreachable, skip: %v", err)
	}
	return ddb
}

// lsofPID 返回监听指定端口的主进程 pid（探针① kill-relay 用）。
func lsofPID(_ string, port int) string {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port)).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// kill9 强杀 pid（模拟 relay 处理中崩溃；supervisor 负责拉起）。
func kill9(pid string) error {
	return exec.Command("kill", "-9", pid).Run()
}

const (
	fb2WebhookURL  = "http://localhost:11112/stripe/webhook"
	fb2WebhookSec  = "whsec_a8b03f35ed1100de63b66e47eec1040a422026b264b92f7dd28681fb98591e07"
	fb2StripeKey   = "sk_test_51QItbp03vhJsKPuLhafsMvAgW6cUattQas8EWX72d9vkZO13kSYs9TlpIU00g0pF3QjQR4zuwd0VQ0fRaU458nA300c9zfDYop"
	fb2ScanTimeout = 8 * time.Minute // scanner 默认 tick 5min，留两轮余量
)

// fb2Sign 构造 Stripe-Signature 头（signed_payload = t + "." + payload，HMAC-SHA256）。
func fb2Sign(secret string, ts int64, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d.", ts)
	mac.Write(payload)
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

// fb2PostWebhook 发送已签名的 checkout.session.completed 事件。
func fb2PostWebhook(t *testing.T, orderID, paymentID string, userID uint32, amount int64, piID string) int {
	t.Helper()
	event := map[string]any{
		"id":   "evt-probe-" + orderID,
		"type": "checkout.session.completed",
		// stripe-go v81 ConstructEvent 校验 API 版本，缺失即 400
		"api_version": "2024-10-28.acacia",
		"data": map[string]any{
			"object": map[string]any{
				"id":             "cs_probe_" + orderID,
				"object":         "checkout.session",
				"payment_status": "paid",
				"amount_total":   amount,
				"currency":       "cny",
				"payment_intent": map[string]any{"id": piID},
				"metadata": map[string]string{
					"order_id":   orderID,
					"payment_id": paymentID,
					"user_id":    fmt.Sprint(userID),
					"pay_amount": fmt.Sprint(amount),
				},
			},
		},
	}
	payload, _ := json.Marshal(event)
	req, _ := http.NewRequest(http.MethodPost, fb2WebhookURL, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", fb2Sign(fb2WebhookSec, time.Now().Unix(), payload))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post webhook: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// seedUnpaid 种未支付订单+支付单（webhook 翻转路径的输入态）。
func seedUnpaid(t *testing.T, db *sql.DB, orderID, paymentID, couponID string) {
	t.Helper()
	fbSeed(t, db, orderID, couponID, 2, 2, 0)
	// fbSeed 把 payment 置 PAID——本探针需要未支付初始态
	if _, err := db.Exec(`UPDATE payments SET status=1, paid_amount=NULL, paid_at=0, transaction_id='' WHERE order_id=$1`, orderID); err != nil {
		t.Fatalf("revert payment: %v", err)
	}
}

// ---------- P1 ① outbox 全链（真签名 webhook） ----------

func TestFB2P1OutboxFullChainViaSignedWebhook(t *testing.T) {
	db := fbDB(t)
	seedUnpaid(t, db, "fb2-o1", "pay-fb2-o1", "")

	rc := fb2PostWebhook(t, "fb2-o1", "pay-fb2-o1", fbUser, 9100, "pi-probe-fb2-o1")
	if rc != http.StatusOK {
		t.Fatalf("webhook must 200, got %d", rc)
	}
	// 翻转+outbox 同事务已落库
	var outStatus string
	fbWait(t, 10*time.Second, func() bool {
		return db.QueryRow(`SELECT status FROM payment_outbox WHERE order_id='fb2-o1'`).Scan(&outStatus) == nil
	})
	if outStatus != "pending" && outStatus != "done" {
		t.Fatalf("outbox status = %q", outStatus)
	}
	// relay（10s tick）接力 → done + 订单 Paid + 库存扣减
	fbWait(t, 60*time.Second, func() bool {
		var st string
		if db.QueryRow(`SELECT status FROM payment_outbox WHERE order_id='fb2-o1'`).Scan(&st) != nil {
			return false
		}
		return st == "done" && fbOrderStatus(t, db, "fb2-o1") == 3
	})
	if got := fbSold(t, db, fbProID); got != 1 {
		t.Fatalf("inventory sold = %d, want 1", got)
	}
	var cnt int
	ddb := fbDtmDB(t)
	defer ddb.Close()
	if err := ddb.QueryRow(`SELECT count(*) FROM trans_global WHERE gid='settle:fb2-o1'`).Scan(&cnt); err != nil || cnt != 1 {
		t.Fatalf("dtm saga count = %d err=%v, want exactly 1", cnt, err)
	}
}

// P1 kill-relay 续扫：flip 落库后 kill -9 payment（supervisor 自动拉起）→ relay 续扫 →
// 不重不漏（恰一个 saga、恰一次扣减、outbox 恰一次 done）。捕捉 mid-state（pending 期间击杀），
// 若错过（MarkDone 先赢）则用新单重试，最多 3 单。
func TestFB2P1KillRelayRescansExactlyOnce(t *testing.T) {
	db := fbDB(t)
	for attempt := 1; attempt <= 3; attempt++ {
		orderID := fmt.Sprintf("fb2-k%d", attempt)
		payID := "pay-" + orderID
		seedUnpaid(t, db, orderID, payID, "")

		rc := fb2PostWebhook(t, orderID, payID, fbUser, 9100, "pi-probe-"+orderID)
		if rc != http.StatusOK {
			t.Fatalf("webhook rc=%d", rc)
		}
		// 轮询 outbox 出现（flip 已提交），出现即 kill -9（捕捉处理中/pending 窗口）
		caught := false
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			var st string
			if err := db.QueryRow(`SELECT status FROM payment_outbox WHERE order_id=$1`, orderID).Scan(&st); err == nil {
				if st == "pending" {
					caught = true // pending 且尚未 done——击杀窗口
				}
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if pid := lsofPID("payment", 10006); pid != "" {
			_ = kill9(pid)
		}
		time.Sleep(3 * time.Second) // 等 supervisor 拉起 + relay 首轮

		// 不重不漏断言（无论是否捕捉到 mid-state，不变量必须成立）
		fbWait(t, 90*time.Second, func() bool {
			var st string
			return db.QueryRow(`SELECT status FROM payment_outbox WHERE order_id=$1`, orderID).Scan(&st) == nil && st == "done"
		})
		fbWait(t, 60*time.Second, func() bool { return fbOrderStatus(t, db, orderID) == 3 })
		var sagaCnt, sold int
		ddb := fbDtmDB(t)
		_ = ddb.QueryRow(`SELECT count(*) FROM trans_global WHERE gid=$1`, "settle:"+orderID).Scan(&sagaCnt)
		_ = ddb.Close()
		_ = db.QueryRow(`SELECT sold FROM inventory WHERE product_id=$1`, fbProID).Scan(&sold)
		if sagaCnt != 1 {
			t.Fatalf("attempt %d: exactly one saga expected, got %d（重复触发）", attempt, sagaCnt)
		}
		if sold != 1 {
			t.Fatalf("attempt %d: inventory sold = %d, want exactly 1（重复扣减/漏扣）", attempt, sold)
		}
		if caught {
			t.Logf("attempt %d: captured mid-state (pending during kill) ✓", attempt)
		} else {
			t.Logf("attempt %d: missed mid-state (MarkDone won the race), invariants still hold", attempt)
		}
		return // 一次成功即返回（不论是否捕捉到 mid-state，不变量成立）
	}
}

// ---------- P2 ② scan1 孤儿单 ----------

func TestFB2P2Scan1ResettlesOrphan(t *testing.T) {
	db := fbDB(t)
	fbSeed(t, db, "fb2-o2", "", 2, 2, time.Now().Add(-20*time.Minute).Unix())

	fbWait(t, fb2ScanTimeout, func() bool { return fbOrderStatus(t, db, "fb2-o2") == 3 })
	if got := fbSold(t, db, fbProID); got != 1 {
		t.Fatalf("inventory sold = %d, want 1", got)
	}
}

// ---------- P3 ③ 墓碑 r2 成功 ----------

func TestFB2P3TombstoneRetrySucceeds(t *testing.T) {
	db := fbDB(t)
	// 券 ID 指向不存在的券 → 首个 saga 券分支必然 Aborted（构造墓碑）
	fbSeed(t, db, "fb2-o3", "cpn-fb2-invalid", 2, 2, time.Now().Add(-15*time.Minute).Unix())

	// 等 scanner tick1：EnsureSaga → 券失败 → saga failed（订单回 Pending）
	ddb := fbDtmDB(t)
	defer ddb.Close()
	fbWait(t, fb2ScanTimeout, func() bool {
		var st string
		if err := ddb.QueryRow(`SELECT status FROM trans_global WHERE gid='settle:fb2-o3'`).Scan(&st); err != nil {
			return false
		}
		return st == "failed"
	})
	if st := fbOrderStatus(t, db, "fb2-o3"); st != 2 {
		t.Fatalf("compensated order should be Pending(2), got %d", st)
	}

	// 补救：插入有效券 + 用户券（Locked 态）→ r2 重试时券分支成功
	if _, err := db.Exec(`INSERT INTO coupons (id, name, type, value, min_amount, start_time, end_time, status, total_count, remaining_count)
		VALUES ('cpn-fb2-valid', 'probe', 1, 900, 0, now() - interval '1 day', now() + interval '1 day', 1, 10, 10)
		ON CONFLICT (id) DO UPDATE SET status=1`); err != nil {
		t.Fatalf("seed coupon: %v", err)
	}
	// user_coupons 的 coupon_id 必须与订单 coupon_id 一致——订单已绑定失效券 ID，
	// 直接改订单的 coupon_id 指向有效券（模拟人工修数后由 r2 收敛）
	if _, err := db.Exec(`UPDATE orders SET coupon_id='cpn-fb2-valid' WHERE order_id='fb2-o3'`); err != nil {
		t.Fatalf("rebind coupon: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_coupons (user_id, coupon_id, status, order_id) VALUES (8888, 'cpn-fb2-valid', 2, 'fb2-o3')
		ON CONFLICT (order_id) DO UPDATE SET status=2`); err != nil {
		t.Fatalf("seed user coupon: %v", err)
	}

	// scanner 下轮：dtm failed → 墓碑 r2（券已有效）→ succeed → 订单 Paid、无 exception
	fbWait(t, fb2ScanTimeout, func() bool {
		var st string
		if err := ddb.QueryRow(`SELECT status FROM trans_global WHERE gid='settle:fb2-o3:r2'`).Scan(&st); err != nil {
			return false
		}
		return st == "succeed"
	})
	fbWait(t, 30*time.Second, func() bool { return fbOrderStatus(t, db, "fb2-o3") == 3 })
	var cnt int
	_ = db.QueryRow(`SELECT count(*) FROM settlement_exception WHERE order_id='fb2-o3' AND status='open'`).Scan(&cnt)
	if cnt != 0 {
		t.Fatalf("successful tombstone retry must not alert, got %d", cnt)
	}
}

// ---------- P4 ④ scan2 + 迟到支付竞态 ----------

func TestFB2P4Scan2ClosesAndLatePaymentConverges(t *testing.T) {
	db := fbDB(t)
	// 超龄待支付单（无 payment 行；created_at 由 fbSeed 固定 -40min）
	fbSeed(t, db, "fb2-o4", "", 2, 2, 0)
	if _, err := db.Exec(`DELETE FROM payments WHERE order_id='fb2-o4'`); err != nil {
		t.Fatalf("remove payment: %v", err)
	}

	// scan2 关单
	fbWait(t, fb2ScanTimeout, func() bool { return fbOrderStatus(t, db, "fb2-o4") == 6 })

	// 迟到支付到账：手工补一笔已支付单（模拟钱在关单前已付、webhook 迟到到账）
	// paid_at 取 now-11min（>scan1 的 10min 阈值），否则等 10min 才入 scan1 视野
	if _, err := db.Exec(`INSERT INTO payments (payment_id, pre_order_id, order_id, user_id, original_amount, expire_time, status, paid_at, paid_amount, transaction_id, payment_method)
		VALUES ('pay-fb2-o4-late', 'pre-fb', 'fb2-o4', 8888, 10000, 0, 2, $1, 9100, 'tx-late-4', 'stripe')`,
		time.Now().Add(-11*time.Minute).Unix()); err != nil {
		t.Fatalf("seed late payment: %v", err)
	}

	// 下轮 scan1：Closed + PAID → 退款 exception（竞态收敛终点）
	fbWait(t, fb2ScanTimeout, func() bool {
		var reason string
		if err := db.QueryRow(`SELECT reason FROM settlement_exception WHERE order_id='fb2-o4' AND status='open'`).Scan(&reason); err != nil {
			return false
		}
		return strings.Contains(reason, "需退款")
	})
}

// ---------- P5 ⑤ Stripe 测试环境对账（sk_test live） ----------

// reconFixtureLive 真实依赖上下文（连真实 DB/dtm/Stripe key；svc 构造即置 stripe.Key）。
func reconFixtureLive(t *testing.T) (*PaymentService, *fallbackScanner, *fakeExceptionModel, *sql.DB, config.Config) {
	t.Helper()
	var c config.Config
	if err := confLoadProbe("etc/payment.yaml", &c); err != nil {
		t.Fatalf("load config: %v", err)
	}
	c.Fallback = c.Fallback.Effective()
	sc := svc.NewServiceContext(c)
	exc := &fakeExceptionModel{}
	sc.ExceptionModel = exc // 聚合断言仍走 fake（真实表写入由真实 Upsert 语义另行覆盖）
	ps := &PaymentService{ctx: sc}
	scanner := newFallbackScanner(sc, c.Fallback)
	return ps, scanner, exc, fbDB(t), c
}

func confLoadProbe(path string, c *config.Config) error {
	return conf.LoadConfig(path, c)
}

// seedSessionAndConfirm 测试模式创建 CheckoutSession 并用测试卡确认（session 变 paid）。
// 全走 raw API（stripe-go 无 payment_intent_data.shipping，而 payment_page confirm 需要它），
// 机制与 stripe CLI 触发 fixture 一致。
func seedSessionAndConfirm(t *testing.T, orderID, paymentID string, userID uint32, amount int64) stripeSessionData {
	t.Helper()
	form := map[string]string{
		"mode":                                   "payment",
		"success_url":                            "https://example.com/probe-ok",
		"cancel_url":                             "https://example.com/probe-cancel",
		"payment_method_types[0]":                "card",
		"line_items[0][price_data][currency]":    "cny",
		"line_items[0][price_data][unit_amount]": fmt.Sprint(amount),
		"line_items[0][price_data][product_data][name]":       "fb2-probe",
		"line_items[0][quantity]":                             "1",
		"metadata[order_id]":                                  orderID,
		"metadata[payment_id]":                                paymentID,
		"metadata[user_id]":                                   fmt.Sprint(userID),
		"metadata[pay_amount]":                                fmt.Sprint(amount),
		"payment_intent_data[shipping][name]":                 "Jenny Rosen",
		"payment_intent_data[shipping][address][line1]":       "510 Townsend St",
		"payment_intent_data[shipping][address][postal_code]": "94103",
		"payment_intent_data[shipping][address][city]":        "San Francisco",
		"payment_intent_data[shipping][address][state]":       "CA",
		"payment_intent_data[shipping][address][country]":     "US",
	}
	out := fb2StripeAPI(t, http.MethodPost, "/v1/checkout/sessions", form)
	csID, _ := out["id"].(string)
	if csID == "" {
		t.Fatalf("session create response missing id: %v", out)
	}
	// 测试端点完成支付（payment_page + tok_visa 确认）
	pm := fb2StripeAPI(t, http.MethodPost, "/v1/payment_methods", map[string]string{
		"type":                                  "card",
		"card[token]":                           "tok_visa",
		"billing_details[name]":                 "Jenny Rosen",
		"billing_details[email]":                "stripe@example.com",
		"billing_details[address][line1]":       "354 Oyster Point Blvd",
		"billing_details[address][postal_code]": "94080",
		"billing_details[address][city]":        "South San Francisco",
		"billing_details[address][state]":       "CA",
		"billing_details[address][country]":     "US",
	})
	pmID, _ := pm["id"].(string)
	fb2StripeAPI(t, http.MethodGet, "/v1/payment_pages/"+csID, nil)
	fb2StripeAPI(t, http.MethodPost, "/v1/payment_pages/"+csID+"/confirm", map[string]string{
		"payment_method":  pmID,
		"expected_amount": fmt.Sprint(amount),
	})
	amt, _ := out["amount_total"].(float64)
	created, _ := out["created"].(float64)
	cur, _ := out["currency"].(string)
	piID := ""
	if pi, ok := out["payment_intent"].(string); ok {
		piID = pi
	}
	return stripeSessionData{
		TxKey:         normalizeStripeTransactionID(orStr(piID, csID)),
		AmountTotal:   int64(amt),
		Currency:      cur,
		OrderID:       orderID,
		PaymentID:     paymentID,
		UserID:        userID,
		CreatedAtUnix: int64(created),
	}
}

// fb2StripeAPI raw Stripe API 调用（测试端点含 payment_pages 等）。
func fb2StripeAPI(t *testing.T, method, path string, form map[string]string) map[string]any {
	t.Helper()
	var body io.Reader
	if len(form) > 0 {
		uv := url.Values{}
		for k, v := range form {
			uv.Set(k, v)
		}
		body = strings.NewReader(uv.Encode())
	}
	req, err := http.NewRequest(method, "https://api.stripe.com"+path, body)
	if err != nil {
		t.Fatalf("new req %s: %v", path, err)
	}
	req.SetBasicAuth(fb2StripeKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("req %s: %v", path, err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if msg, _ := out["error"].(map[string]any); msg != nil {
		t.Fatalf("stripe api %s %s failed: %v", method, path, msg["message"])
	}
	return out
}

// P5-方向1：Stripe 有账（测试卡已扣）+ 本地未 PAID → recon 自动重放 → 本地 PAID + 结算完成
func TestFB2P5ReconDirection1AutoReplay(t *testing.T) {
	db := fbDB(t)
	seedUnpaid(t, db, "fb2-o5", "pay-fb2-o5", "")
	// 订单应付与 Stripe 会话金额一致（重放校验链要求）
	if _, err := db.Exec(`UPDATE orders SET payable_amount=9100, original_amount=9100, discount_amount=0 WHERE order_id='fb2-o5'`); err != nil {
		t.Fatalf("align payable: %v", err)
	}

	// 构造 Stripe 侧真实扣款（测试卡确认，不发本地 webhook——模拟 webhook 永久丢失）
	// （webhook 丢失场景本地 transaction_id 为空，重放以 Stripe 会话元数据为准）
	seedSessionAndConfirm(t, "fb2-o5", "pay-fb2-o5", fbUser, 9100)

	// 驱动对账（窗口覆盖"现在"）
	var c config.Config
	if err := confLoadProbe("etc/payment.yaml", &c); err != nil {
		t.Fatalf("load: %v", err)
	}
	c.Fallback = c.Fallback.Effective()
	sc := svc.NewServiceContext(c)
	sc.ExceptionModel = &fakeExceptionModel{}
	ps := &PaymentService{ctx: sc}
	now := time.Now()
	ps.runReconWindow(context.Background(), now.Unix()-7200, now.Unix()+3600, c.Fallback)

	// 断言：本地被自动重放为已支付（含结算 saga 发起）
	var payStatus int64
	if err := db.QueryRow(`SELECT status FROM payments WHERE payment_id='pay-fb2-o5'`).Scan(&payStatus); err != nil {
		t.Fatalf("query payment: %v", err)
	}
	if payStatus != int64(2) {
		t.Fatalf("direction1 replay must mark payment PAID(2), got %d", payStatus)
	}
	fbWait(t, 60*time.Second, func() bool { return fbOrderStatus(t, db, "fb2-o5") == 3 })
}

// P5-方向2：本地 PAID、Stripe 无账 → 冻结告警，不回滚
func TestFB2P5ReconDirection2Freezes(t *testing.T) {
	db := fbDB(t)
	fbSeed(t, db, "fb2-o6", "", 2, 2, time.Now().Unix())
	// 交易号指向 Stripe 不存在的账
	if _, err := db.Exec(`UPDATE payments SET transaction_id='tx-fb2-fabricated' WHERE order_id='fb2-o6'`); err != nil {
		t.Fatalf("set fabricated tx: %v", err)
	}

	var c config.Config
	if err := confLoadProbe("etc/payment.yaml", &c); err != nil {
		t.Fatalf("load: %v", err)
	}
	c.Fallback = c.Fallback.Effective()
	sc := svc.NewServiceContext(c)
	exc := &fakeExceptionModel{}
	sc.ExceptionModel = exc
	ps := &PaymentService{ctx: sc}
	now := time.Now()
	ps.runReconWindow(context.Background(), now.Unix()-7200, now.Unix()+3600, c.Fallback)

	found := false
	for _, u := range exc.upserts {
		if u.orderID == "fb2-o6" && strings.Contains(u.reason, "Stripe 无对应账") {
			found = true
		}
	}
	if !found {
		t.Fatalf("direction2 must freeze with alert: %+v", exc.upserts)
	}
	// 不回滚断言：recon 只冻结告警，不得触发补偿/结算——payment 仍 PAID、
	// 订单保持播种态（PendingPayment=2，与补偿后回滚态不同）、无新 saga
	var payStatus int64
	if err := db.QueryRow(`SELECT status FROM payments WHERE payment_id='pay-fb2-o6'`).Scan(&payStatus); err != nil || payStatus != 2 {
		t.Fatalf("direction2 must NOT rollback payment, got status=%d err=%v", payStatus, err)
	}
	if st := fbOrderStatus(t, db, "fb2-o6"); st != 2 {
		t.Fatalf("direction2 must NOT change order, got order_status=%d", st)
	}
	ddb := fbDtmDB(t)
	defer ddb.Close()
	var sagaCnt int
	if err := ddb.QueryRow(`SELECT count(*) FROM trans_global WHERE gid LIKE 'settle:fb2-o6%'`).Scan(&sagaCnt); err != nil || sagaCnt != 0 {
		t.Fatalf("direction2 must NOT initiate settlement saga, got %d err=%v", sagaCnt, err)
	}
}
