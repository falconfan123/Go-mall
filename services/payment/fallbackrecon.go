package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/stripe/stripe-go/v81"
	session "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/zeromicro/go-zero/core/logx"
)

// reconLocalPaid 对账窗口内的本地已支付单（方向2 集合）。
type reconLocalPaid struct {
	PaymentID     string
	OrderID       string
	UserID        int64
	TransactionID string
}

// stripeSessionData 对账所需的 Stripe 会话快照（测试注入的抽象边界，Rule 6）。
type stripeSessionData struct {
	TxKey         string // normalizeStripeTransactionID(PaymentIntent.ID 优先, 其次 session.ID)
	AmountTotal   int64
	Currency      string
	OrderID       string
	PaymentID     string
	UserID        uint32
	CreatedAtUnix int64
}

// defaultStripeLister 生产实现：Stripe CheckoutSessionList 分页拉取。
func defaultStripeLister(ctx context.Context, startSec, endSec int64) ([]stripeSessionData, error) {
	params := &stripe.CheckoutSessionListParams{
		CreatedRange: &stripe.RangeQueryParams{
			GreaterThanOrEqual: startSec,
			LesserThan:         endSec,
		},
		Status: stripe.String("complete"),
	}
	params.Limit = stripe.Int64(100)
	var out []stripeSessionData
	iter := session.List(params)
	for iter.Next() {
		cs := iter.CheckoutSession()
		if cs.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
			continue
		}
		var piID string
		if cs.PaymentIntent != nil && cs.PaymentIntent.ID != "" {
			piID = cs.PaymentIntent.ID
		}
		out = append(out, stripeSessionData{
			TxKey:         normalizeStripeTransactionID(orStr(piID, cs.ID)),
			AmountTotal:   cs.AmountTotal,
			Currency:      string(cs.Currency),
			OrderID:       cs.Metadata["order_id"],
			PaymentID:     cs.Metadata["payment_id"],
			UserID:        parseUintOrZero(cs.Metadata["user_id"]),
			CreatedAtUnix: cs.Created,
		})
	}
	return out, iter.Err()
}

func orStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func parseUintOrZero(s string) uint32 {
	v, _ := strconv.ParseUint(s, 10, 32)
	return uint32(v)
}

// reconOnce 治⑤⑥：Stripe 双向日终对账（design 决策 8）。
// 每小时 tick：仅在 ReconHour 且当日未跑时执行（当日幂等）。
func (s *PaymentService) reconOnce(ctx context.Context, fb config.FallbackConfig) {
	now := time.Now()
	if now.Hour() != fb.ReconHour {
		return
	}
	today := now.Format("2006-01-02")
	s.reconMu.Lock()
	if s.lastReconDay == today {
		s.reconMu.Unlock()
		return
	}
	s.lastReconDay = today // 先占位防并发双跑；失败会在次日重试（对账幂等聚合）
	s.reconMu.Unlock()

	// T-1 日窗（本地时区）
	dayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	logx.Infow("stripe reconciliation scheduled for T-1 window",
		logx.Field("window", dayStart.Format(time.DateTime)+"~"+dayEnd.Format(time.DateTime)))
	s.runReconWindow(ctx, dayStart.Unix(), dayEnd.Unix(), fb)
}

// runReconWindow 对指定时间窗执行双向对账（reconOnce 的窗口执行体；探针可注入任意窗口）。
func (s *PaymentService) runReconWindow(ctx context.Context, startSec, endSec int64, fb config.FallbackConfig) {
	startSec -= 3600 // Stripe 侧窗口 ±1h 吸收回调延迟
	endSec += 3600
	logx.Infow("stripe reconciliation start",
		logx.Field("window_start", startSec), logx.Field("window_end", endSec))

	sessions, err := s.listStripeSessions(ctx, startSec-3600, endSec+3600)
	if err != nil {
		logx.Errorw("stripe reconciliation list failed (will retry next day)",
			logx.Field("err", err))
		return
	}

	stripeByTx := map[string]stripeSessionData{}
	for _, ss := range sessions {
		if ss.TxKey == "" {
			continue
		}
		stripeByTx[ss.TxKey] = ss
	}

	// 本地窗口内已支付单（方向2 的本地集合）
	var local []reconLocalPaid
	localQuery := `SELECT payment_id, order_id, user_id, transaction_id FROM payments
		WHERE status = 2 AND paid_at >= $1 AND paid_at < $2`
	if err := s.ctx.RawConn.QueryRowsCtx(ctx, &local, localQuery, startSec, endSec); err != nil {
		logx.Errorw("stripe reconciliation local query failed", logx.Field("err", err))
		return
	}

	localByTx := map[string]reconLocalPaid{}
	for _, lp := range local {
		if lp.TransactionID == "" {
			continue
		}
		localByTx[strings.ToLower(lp.TransactionID)] = lp
	}

	// 方向1：Stripe 有、本地未 PAID（或缺失）→ 校验金额/币种 → 自动重放
	for txKey, ss := range stripeByTx {
		lowerKey := strings.ToLower(txKey)
		lp, exists := localByTx[lowerKey]
		if exists && lp.OrderID != "" {
			var status int64
			qerr := s.ctx.RawConn.QueryRowCtx(ctx, &status,
				`SELECT status FROM payments WHERE payment_id = $1`, lp.PaymentID)
			if qerr == nil && status == int64(2) {
				continue // 本地已 PAID，正常
			}
		}
		// 本地未 PAID（或无支付单）→ 重放前校验
		if ss.Currency != "cny" {
			s.recordReconException(ctx, ss, "币种不一致，停止自动重放",
				map[string]any{"currency": ss.Currency, "amount_total": ss.AmountTotal})
			continue
		}
		if ss.OrderID == "" || ss.PaymentID == "" || ss.UserID == 0 {
			s.recordReconException(ctx, ss, "Stripe 会话元数据缺失，停止自动重放", nil)
			continue
		}
		var payable int64
		perr := s.ctx.RawConn.QueryRowCtx(ctx, &payable,
			`SELECT payable_amount FROM orders WHERE order_id = $1`, ss.OrderID)
		if perr != nil || payable != ss.AmountTotal {
			s.recordReconException(ctx, ss, "金额与本地应付不一致，停止自动重放",
				map[string]any{"stripe_amount": ss.AmountTotal, "payable_amount": payable, "query_err": errStr(perr)})
			continue
		}
		// 自动重放（安全垫：支付状态机 + EnsureSaga 幂等 + outbox + 扫描）
		if err := s.processStripePaymentSuccess(ctx, ss.OrderID, ss.PaymentID, ss.UserID,
			ss.TxKey, ss.AmountTotal, ss.CreatedAtUnix); err != nil {
			s.recordReconException(ctx, ss, "丢单重放失败，转人工",
				map[string]any{"err": err.Error()})
			continue
		}
		logx.Infow("stripe reconciliation replayed lost payment", logx.Field("order_id", ss.OrderID))
	}

	// 方向2：本地 PAID、Stripe 无账 → 告警冻结，不自动回滚
	for _, lp := range local {
		key := strings.ToLower(lp.TransactionID)
		if key == "" {
			s.recordReconException(ctx, stripeSessionData{OrderID: lp.OrderID, PaymentID: lp.PaymentID},
				"已支付但交易号为空", nil)
			continue
		}
		if _, ok := stripeByTx[key]; !ok {
			s.recordReconException(ctx, stripeSessionData{OrderID: lp.OrderID, PaymentID: lp.PaymentID},
				"本地已支付但 Stripe 无对应账，冻结待查",
				map[string]any{"transaction_id": lp.TransactionID})
		}
	}
	logx.Infow("stripe reconciliation done",
		logx.Field("stripe_sessions", len(stripeByTx)), logx.Field("local_paid", len(local)))
}

func (s *PaymentService) listStripeSessions(ctx context.Context, startSec, endSec int64) ([]stripeSessionData, error) {
	if s.stripeLister != nil {
		return s.stripeLister(ctx, startSec, endSec)
	}
	return defaultStripeLister(ctx, startSec, endSec)
}

func (s *PaymentService) recordReconException(ctx context.Context, ss stripeSessionData, reason string, detail map[string]any) {
	recordException(ctx, s.ctx, "recon", ss.OrderID, ss.PaymentID, reason, detail)
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
