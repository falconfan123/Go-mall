package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
)

// runLoop 周期任务框架（design 决策 1）：
//   - 启动后立即执行首轮，随后按 tick 周期触发；
//   - 每轮 panic recover（error 日志后本轮回收，循环不死）；
//   - ctx 取消后当前轮跑完即退出（不立即中断，design 决策 1/9 的优雅停机语义）。
func runLoop(ctx context.Context, tick time.Duration, name string, fn func(context.Context)) {
	run := func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorw("fallback loop tick panicked",
					logx.Field("loop", name), logx.Field("panic", r))
			}
		}()
		fn(ctx)
	}
	run() // 首轮立即执行
	timer := time.NewTicker(tick)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			logx.Infow("fallback loop stopped by context",
				logx.Field("loop", name))
			return
		case <-timer.C:
			run()
		}
	}
}

// startFallbackWorker 启动结算兜底三层常驻循环（design 决策 1）：
// relay（outbox 接力）/ scanner（对账扫描）/ recon（Stripe 日终对账）。
// 生命周期：ctx 挂 AddWrapUpListener，服务停机时 cancel，各循环当前轮跑完退出。
func (s *PaymentService) startFallbackWorker() {
	fb := s.ctx.Config.Fallback.Effective()
	ctx, cancel := context.WithCancel(context.Background())
	proc.AddWrapUpListener(cancel)

	go runLoop(ctx, time.Duration(fb.RelayTickMs)*time.Millisecond, "relay", s.relayOnce)
	go runLoop(ctx, time.Duration(fb.ScannerTickMs)*time.Millisecond, "scanner", s.scannerOnce)
	if fb.ReconDisabled {
		logx.Infow("stripe reconciliation disabled by config")
		return
	}
	go runLoop(ctx, time.Hour, "recon", func(c context.Context) { s.reconOnce(c, fb) })
}

// relayOnce 治②：outbox pending 待办接力（design 决策 5）。
// 认领（条件更新租约）→ EnsureSaga → 成功即 done（不等待 saga 终态）；
// 失败指数退避重试，超限置 dead + exception(source=relay)。
func (s *PaymentService) relayOnce(ctx context.Context) {
	fb := s.ctx.Config.Fallback.Effective()
	batch, err := s.ctx.OutboxModel.FetchPendingBatch(ctx, 50)
	if err != nil {
		logx.Errorw("relay fetch pending failed", logx.Field("err", err))
		return
	}
	for _, item := range batch {
		// 并发安全认领：条件更新推进租约；崩溃后租约到期自然重新可见
		claimed, err := s.ctx.OutboxModel.ClaimPending(ctx, item.Id, 60)
		if err != nil {
			logx.Errorw("relay claim failed", logx.Field("err", err), logx.Field("id", item.Id))
			continue
		}
		if !claimed {
			continue
		}

		var input svc.SettlementInput
		if err := json.Unmarshal([]byte(item.Payload), &input); err != nil {
			// payload 损坏 = 毒待办：重试无意义，直接 dead + 人工产出
			_ = s.ctx.OutboxModel.MarkDead(ctx, item.Id, "payload decode failed: "+err.Error())
			recordException(ctx, s.ctx, "relay", item.OrderId, item.PaymentId, "outbox payload decode failed", nil)
			continue
		}

		if err := s.ctx.SettlementSaga.Ensure(ctx, &input); err != nil {
			// 指数退避：2^attempts × relay tick 秒
			backoff := (fb.RelayTickMs / 1000) << uint(item.Attempts)
			attempts, aerr := s.ctx.OutboxModel.IncrAttempt(ctx, item.Id, backoff)
			if aerr != nil {
				logx.Errorw("relay incr attempt failed", logx.Field("err", aerr), logx.Field("id", item.Id))
			}
			if aerr == nil && attempts >= int64(fb.RelayMaxAttempts) {
				_ = s.ctx.OutboxModel.MarkDead(ctx, item.Id, err.Error())
				recordException(ctx, s.ctx, "relay", item.OrderId, item.PaymentId,
					"outbox retry limit exceeded",
					map[string]any{"err": err.Error(), "attempts": attempts})
				continue
			}
			logx.Infow("outbox item requeued for retry",
				logx.Field("id", item.Id), logx.Field("order_id", item.OrderId),
				logx.Field("attempt", attempts), logx.Field("cause", err.Error()))
			continue
		}

		// done 语义：EnsureSaga 提交成功即 done，不等待 saga 终态（C3）
		if err := s.ctx.OutboxModel.MarkDone(ctx, item.Id); err != nil {
			logx.Errorw("relay mark done failed", logx.Field("err", err), logx.Field("id", item.Id))
		}
		logx.Infow("outbox item ensured", logx.Field("id", item.Id), logx.Field("order_id", item.OrderId))
	}
}

// recordException 异常单人工产出（design 决策 2）：表聚合去重 + 需人工介入级告警；
// isNew=false（重复发现）时不再重复触发告警。
func recordException(ctx context.Context, sc *svc.ServiceContext, source, orderID, paymentID, reason string, detail map[string]any) {
	seenCount, isNew, err := sc.ExceptionModel.UpsertOpen(ctx, orderID, paymentID, source, reason, detail)
	if err != nil {
		logx.Errorw("record exception failed",
			logx.Field("err", err), logx.Field("order_id", orderID), logx.Field("source", source))
		return
	}
	logx.Errorw("settlement exception recorded",
		logx.Field("order_id", orderID), logx.Field("source", source),
		logx.Field("reason", reason), logx.Field("seen_count", seenCount),
		logx.Field("is_new", isNew))
}
