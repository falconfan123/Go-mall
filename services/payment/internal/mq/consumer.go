package mq

import (
	"context"
	"encoding/json"
	"time"

	idempotency "github.com/falconfan123/Go-mall/common/utils/idempotency"
	"github.com/zeromicro/go-zero/core/logx"
)

func (a *PaymentDelayMQ) consumer(ctx context.Context) {

	ch, err := a.conn.Channel()
	if err != nil {
		logx.Errorw("Failed to open a channel", logx.Field("err", err))
		return
	}

	results, err := ch.Consume(
		QueueName, // 队列名称
		"",        // 消费者标签
		true,      // 自动确认（ack）
		false,     // 排他性
		false,     // 本地消息
		false,     // 等待确认
		nil,       // 参数
	)
	if err != nil {
		logx.Errorw("Failed to register a consumer", logx.Field("err", err))
	}
	logx.Infow("Starting RabbitMQ consumer...")

	for res := range results {
		var msg *PaymentReq
		if err := json.Unmarshal(res.Body, &msg); err != nil {
			logx.Errorw("failed to unmarshal message", logx.Field("error", err), logx.Field("body", string(res.Body)))
			if err := res.Reject(false); err != nil {
				logx.Errorw("failed to reject message", logx.Field("error", err), logx.Field("body", string(res.Body)))
			}
			continue
		}

		// 幂等检查：检查并设置去重 key
		idempotencyKey := idempotency.BuildKey("payment", "delay", msg.OrderId)
		isProcessed, err := idempotency.CheckAndSet(ctx, a.Redis, idempotencyKey, time.Hour)
		if err != nil {
			logx.Errorw("failed to check idempotency", logx.Field("err", err), logx.Field("key", idempotencyKey))
			if err := res.Reject(true); err != nil {
				logx.Errorw("failed to reject message", logx.Field("error", err), logx.Field("body", string(res.Body)))
			}
			continue
		}
		if isProcessed {
			logx.Infow("message already processed, skipping", logx.Field("key", idempotencyKey))
			if err := res.Ack(false); err != nil {
				logx.Errorw("failed to ack message", logx.Field("error", err), logx.Field("body", string(res.Body)))
			}
			continue
		}

	}
}
