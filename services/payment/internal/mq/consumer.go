package mq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/falconfan123/Go-mall/common/mq/consumerx"
)

// paymentDelayHandler 是 payment-delay 队列的业务回调。
// 当前为空实现扩展点：队列在 svc 中保持禁用（决策 7，消除带病休眠代码，
// 复活时业务逻辑在此补全）。骨架能力与 order 侧 consumer 一致。
type paymentDelayHandler struct{}

func (h *paymentDelayHandler) Decode(body []byte) (*PaymentReq, error) {
	msg := &PaymentReq{}
	if err := json.Unmarshal(body, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (h *paymentDelayHandler) IdempotencyKey(msg *PaymentReq) string {
	return msg.OrderId
}

func (h *paymentDelayHandler) Handle(context.Context, *PaymentReq) error {
	// 扩展点：支付超时对账/关单逻辑待复活该队列时实现（独立变更）
	return nil
}

// consumer 用加固骨架消费 payment-delay 队列：
// autoAck=false 手动确认、监督重启、幂等 claim/release、有界重试、毒消息一次即弃。
func (a *PaymentDelayMQ) consumer(ctx context.Context) {
	c := consumerx.New[*PaymentReq](
		a.conn,
		a.dial,
		consumerx.NewRedisStore(a.Redis),
		&paymentDelayHandler{},
		consumerx.Config{
			QueueName:          QueueName,
			IdempotencyService: "payment",
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
func (a *PaymentDelayMQ) Start(ctx context.Context) {
	go a.consumer(ctx)
}
