package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	inventorypb "github.com/falconfan123/Go-mall/services/inventory/pb"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// scan2Row 扫描行（orders 超龄待支付）。
type scan2Row struct {
	OrderID string `db:"order_id"`
	UserID  int64  `db:"user_id"`
}

// fallbackScanner 定时对账扫描（design 决策 6，治③④）：
// scan1：钱到了货没跟上（分流：PendingPayment 补结算 / 关闭+取消 转人工产出）；
// scan2：该关没关（复用 order 关单入口）；同一订单跨层发现由 exception 表聚合去重。
type fallbackScanner struct {
	ctx *svc.ServiceContext
	cfg config.FallbackConfig
	// scan2 连续失败计数（进程内存；重启重置为更保守方向——仅延迟告警，可接受）
	scan2FailsMu sync.Mutex
	scan2Fails   map[string]int
}

func newFallbackScanner(sc *svc.ServiceContext, fb config.FallbackConfig) *fallbackScanner {
	return &fallbackScanner{ctx: sc, cfg: fb, scan2Fails: map[string]int{}}
}

func (s *PaymentService) scannerOnce(ctx context.Context) {
	sc := newFallbackScanner(s.ctx, s.ctx.Config.Fallback.Effective())
	sc.scan1(ctx)
	sc.scan2(ctx)
}

// scan1Row 扫描行（payments ⨝ orders）。
type scan1Row struct {
	OrderID        string `db:"order_id"`
	UserID         int64  `db:"user_id"`
	TransactionID  string `db:"transaction_id"`
	PaidAmount     int64  `db:"paid_amount"`
	PaidAt         int64  `db:"paid_at"`
	OrderStatus    int64  `db:"order_status"`
	PreOrderID     string `db:"pre_order_id"`
	CouponID       string `db:"coupon_id"`
	OriginalAmount int64  `db:"original_amount"`
	DiscountAmount int64  `db:"discount_amount"`
}

// scan1Item 订单行聚合。
type scan1Item struct {
	ProductId int32 `db:"product_id"`
	Quantity  int32 `db:"quantity"`
}

// dtmSagaStatus 查询 saga 状态（dtm HTTP query API，v1.19.0 返回 {transaction:{status},branches}）。
// 返回 "" 表示无记录（gid 未使用）。
func dtmSagaStatus(ctx context.Context, httpAddr, gid string) (string, error) {
	if httpAddr == "" {
		return "", fmt.Errorf("dtm http addr not configured")
	}
	if !strings.HasPrefix(httpAddr, "http://") && !strings.HasPrefix(httpAddr, "https://") {
		httpAddr = "http://" + httpAddr
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		httpAddr+"/api/dtmsvr/query?gid="+url.QueryEscape(gid), nil)
	if err != nil {
		return "", err
	}
	// 显式超时：dtm 慢响应不得阻塞 scanner 后续 scan2（实施期发现）
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Transaction *struct {
			Status string `json:"status"`
		} `json:"transaction"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Transaction == nil {
		return "", nil
	}
	return out.Transaction.Status, nil
}

func (s *fallbackScanner) scan1(ctx context.Context) {
	var rows []scan1Row
	// 时间比较：paid_at(bigint, unix 秒) 对 floor(epoch)::bigint，参数化阈值防时区坑
	query := `SELECT p.order_id, p.user_id, p.transaction_id, p.paid_amount, p.paid_at,
		o.order_status, o.pre_order_id, o.coupon_id, o.original_amount, o.discount_amount
		FROM payments p JOIN orders o ON o.order_id = p.order_id
		WHERE p.status = 2
		  AND p.paid_at < floor(extract(epoch from (now() - ($1::text || ' minutes')::interval)))::bigint
		  AND o.order_status <> 3
		LIMIT 100`
	if err := s.ctx.RawConn.QueryRowsCtx(ctx, &rows, query,
		fmt.Sprint(s.cfg.Scan1ThresholdMinutes)); err != nil {
		logx.Errorw("scan1 query failed", logx.Field("err", err))
		return
	}
	for i := range rows {
		s.scan1Row(ctx, &rows[i])
	}
}

func (s *fallbackScanner) scan1Row(ctx context.Context, r *scan1Row) {
	input := &svc.SettlementInput{
		OrderID:        r.OrderID,
		UserID:         int32(r.UserID),
		PreOrderID:     r.PreOrderID,
		TransactionID:  r.TransactionID,
		PaidAmount:     r.PaidAmount,
		PaidAt:         r.PaidAt,
		CouponID:       r.CouponID,
		DiscountAmount: r.DiscountAmount,
		OriginAmount:   r.OriginalAmount,
	}

	switch r.OrderStatus {
	case int64(2): // PendingPayment → 补结算
		var items []scan1Item
		if err := s.ctx.RawConn.QueryRowsCtx(ctx, &items,
			`SELECT product_id, quantity FROM order_items WHERE order_id = $1`, r.OrderID); err != nil {
			logx.Errorw("scan1 query order items failed",
				logx.Field("err", err), logx.Field("order_id", r.OrderID))
			return
		}
		for _, it := range items {
			input.Items = append(input.Items, &inventorypb.InventoryReq_Items{
				ProductId: it.ProductId,
				Quantity:  it.Quantity,
			})
		}
		s.scan1Resettle(ctx, input)
	case int64(5), int64(6): // Cancelled / Closed → 补偿语义正确，转人工（退款另立项）
		s.recordScanException(ctx, "scan1", r.OrderID, r.TransactionID,
			"已支付订单已关闭/取消，需退款", map[string]any{"order_status": r.OrderStatus})
	default:
		logx.Infow("scan1 skip unknown order status", logx.Field("order_status", r.OrderStatus))
	}
}

// scan1Resettle 分流补结算：无 saga → EnsureSaga；in-flight → 干等；failed（墓碑）→
// 递增后缀 gid 有限重试（上限 3）；succeed → 数据矛盾 exception。
func (s *fallbackScanner) scan1Resettle(ctx context.Context, input *svc.SettlementInput) {
	status, err := dtmSagaStatus(ctx, s.ctx.Config.Dtm.HttpAddr, "settle:"+input.OrderID)
	if err != nil {
		// dtm 不可达：保守跳过本轮 + 日志，不误判（下轮再试）
		logx.Errorw("scan1 dtm query failed, skip this round",
			logx.Field("err", err), logx.Field("order_id", input.OrderID))
		return
	}
	switch status {
	case "":
		// 无 saga 记录（洞②漏网）→ 幂等补建
		if err := s.ctx.SettlementSaga.Ensure(ctx, input); err != nil {
			logx.Errorw("scan1 ensure failed", logx.Field("err", err), logx.Field("order_id", input.OrderID))
		}
		return
	case "submitted", "prepared":
		// in-flight：编排层在重试，干等不干预（spec："进行中 Saga 干等不干预"）
		logx.Infow("scan1 saga in-flight, wait",
			logx.Field("order_id", input.OrderID), logx.Field("status", status))
		return
	case "succeed":
		// 数据矛盾（订单非 Paid 却 succeed，理论不可达）→ 按"异常单人工产出"处理
		s.recordScanException(ctx, "scan1", input.OrderID, input.TransactionID,
			"数据矛盾：saga succeed 但订单未 Paid", map[string]any{"saga_status": status})
		return
	case "failed":
		// 墓碑：递增后缀 gid 有限自动重试
		s.tombstoneRetry(ctx, input)
		return
	default:
		logx.Infow("scan1 unknown saga status",
			logx.Field("order_id", input.OrderID), logx.Field("status", status))
	}
}

// tombstoneRetry 依次探测原始 gid 与 r2..r{n}：找到未用后缀即发起新 saga；
// 全部已用且 failed → exception 转人工。
func (s *fallbackScanner) tombstoneRetry(ctx context.Context, input *svc.SettlementInput) {
	for n := 2; n <= s.cfg.TombstoneMaxRetry; n++ {
		suffix := fmt.Sprintf(":r%d", n)
		gid := "settle:" + input.OrderID + suffix
		status, err := dtmSagaStatus(ctx, s.ctx.Config.Dtm.HttpAddr, gid)
		if err != nil {
			logx.Errorw("tombstone probe failed", logx.Field("err", err), logx.Field("gid", gid))
			return
		}
		if status == "" {
			retry := *input
			retry.GidSuffix = suffix
			if err := s.ctx.SettlementSaga.Initiate(ctx, &retry); err != nil {
				logx.Errorw("tombstone retry initiate failed",
					logx.Field("err", err), logx.Field("gid", gid))
			} else {
				logx.Infow("tombstone retry initiated", logx.Field("gid", gid))
			}
			return
		}
		if status == "submitted" || status == "prepared" {
			logx.Infow("tombstone retry in-flight, wait", logx.Field("gid", gid))
			return
		}
		if status == "succeed" {
			s.recordScanException(ctx, "scan1", input.OrderID, input.TransactionID,
				"数据矛盾：墓碑重试 succeed 但订单未 Paid", map[string]any{"gid": gid})
			return
		}
		// failed → 继续探测下一个后缀
	}
	s.recordScanException(ctx, "scan1", input.OrderID, input.TransactionID,
		"墓碑重试超限（r2..r{n} 均 failed）",
		map[string]any{"max_retry": s.cfg.TombstoneMaxRetry})
}

// scan2 该关没关：超龄待支付订单 → 复用 order 关单入口（RPC）；与迟到支付的竞态由
// 既有状态机语义收敛（关单后钱到 → scan1 下轮 关闭+PAID 分支产出退款待办）。
func (s *fallbackScanner) scan2(ctx context.Context) {
	var rows []scan2Row
	query := `SELECT order_id, user_id FROM orders
		WHERE order_status = 2 AND payment_status = 2
		  AND created_at < now() - ($1::text || ' minutes')::interval
		LIMIT 100`
	if err := s.ctx.RawConn.QueryRowsCtx(ctx, &rows, query,
		fmt.Sprint(s.cfg.Scan2ThresholdMinutes)); err != nil {
		logx.Errorw("scan2 query failed", logx.Field("err", err))
		return
	}
	for _, r := range rows {
		_, err := s.ctx.OrderRpc.CloseExpiredOrder(ctx, &orderpb.CloseExpiredOrderRequest{
			OrderId: r.OrderID,
			UserId:  int32(r.UserID),
		})
		if err != nil {
			// 连续失败计数；超阈值产出 exception 并告警，本单仍继续尝试（不限次）
			s.scan2FailsMu.Lock()
			s.scan2Fails[r.OrderID]++
			fails := s.scan2Fails[r.OrderID]
			s.scan2FailsMu.Unlock()
			if fails >= 3 {
				s.recordScanException(ctx, "scan2", r.OrderID, "",
					"关单连续失败（超龄单）",
					map[string]any{"err": err.Error(), "consecutive_failures": fails})
			} else {
				logx.Infow("scan2 close failed, will retry next tick",
					logx.Field("order_id", r.OrderID),
					logx.Field("cause", err.Error()), logx.Field("fails", fails))
			}
			continue
		}
		s.scan2FailsMu.Lock()
		delete(s.scan2Fails, r.OrderID)
		s.scan2FailsMu.Unlock()
		logx.Infow("scan2 closed expired order", logx.Field("order_id", r.OrderID))
	}
}

// recordScanException 表聚合（isNew=false 不重复告警）+ 需人工介入级告警。
func (s *fallbackScanner) recordScanException(ctx context.Context, source, orderID, paymentID, reason string, detail map[string]any) {
	recordException(ctx, s.ctx, source, orderID, paymentID, reason, detail)
}
