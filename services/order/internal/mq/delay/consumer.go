package delay

import (
	"context"
	"errors"
	"time"

	"github.com/falconfan123/Go-mall/common/mq/consumerx"
)

// errSkipSideEffects 哨兵错误：订单状态非 Created，跳过后续副作用（按成功路径确认）。
var errSkipSideEffects = errors.New("skip side effects")

// consumer 用加固骨架消费 order-delay 死信队列：
// autoAck=false 手动确认、监督重启、幂等 claim/release、有界重试、毒消息一次即弃。
func (a *OrderDelayMQ) consumer(ctx context.Context) {
	c := consumerx.New[*OrderReq](
		a.conn,
		a.dial,
		consumerx.NewRedisStore(a.Redis),
		newDelayHandler(a),
		consumerx.Config{
			QueueName:          DeadLetterQueue,
			IdempotencyService: "order",
			IdempotencyQueue:   "delay",
			RetryLimit:         a.consumerCfg.RetryLimit,
			BackoffBase:        time.Duration(a.consumerCfg.BackoffBaseMs) * time.Millisecond,
			ShutdownTimeout:    time.Duration(a.consumerCfg.ShutdownTimeoutMs) * time.Millisecond,
			IdempotencyTTL:     time.Duration(a.consumerCfg.IdempotencyTtlSeconds) * time.Second,
			SuccessTTL:         time.Duration(a.consumerCfg.SuccessTtlSeconds) * time.Second,
		},
	)
	c.Run(ctx)
}

// Start 在独立 goroutine 中启动消费者；ctx 取消时优雅退出（服务停机时触发）。
func (a *OrderDelayMQ) Start(ctx context.Context) {
	go a.consumer(ctx)
}
