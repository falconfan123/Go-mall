// Package consumerx 提供结算链路 MQ consumer 的加固骨架：
// 手动确认（autoAck=false）、监督重启（指数退避）、幂等 claim/release、
// 有界重试（超限弃单告警）、毒消息一次即弃、panic recovery、优雅停机。
//
// 骨架包办：通道建立与 QoS、消费循环、JSON 解码入口、幂等 claim/release、
// 重试计数与上限判定、Ack/Reject 原语、panic recovery、全部结构化日志、监督重启。
// 业务方实现 Handler[T] 注入：Decode / IdempotencyKey / Handle。
package consumerx

import (
	"context"
	"errors"
	"time"

	"github.com/falconfan123/Go-mall/common/utils/idempotency"
	"github.com/streadway/amqp"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	maxBackoff     = 30 * time.Second
	retryKeyPrefix = "settlement:retry"

	defaultRetryLimit        = 3
	defaultBackoffBase       = time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultIdempotencyTTL    = time.Hour
	defaultSuccessTTL        = 24 * time.Hour
	defaultConnectionTimeout = 5 * time.Second
)

// Store 抽象幂等/重试计数所需的 Redis 操作（两态键语义，design 决策 4），便于单测注入 fake。
type Store interface {
	// Claim 两态占坑：返回状态（New=占坑成功可处理 / Processing=在途跳过 /
	// Done=已成功跳过）。租约过期后 key 自动消失，再次 Claim 即 New。
	Claim(ctx context.Context, key string, leaseTTL time.Duration) (idempotency.ClaimState, error)
	// MarkDone 全部业务成功后调用：将标记覆盖为 SUCCESS（PX=successTTL）。
	MarkDone(ctx context.Context, key string, successTTL time.Duration) error
	// Release 释放占坑 key（失败路径），使重投消息可被正常处理。
	Release(ctx context.Context, key string) error
	// IncrWithTTL 重试计数自增并刷新 TTL，返回自增后的值。
	IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type redisStore struct {
	rds *redis.Redis
}

// NewRedisStore 将 go-zero Redis 适配为 Store。
func NewRedisStore(rds *redis.Redis) Store {
	return redisStore{rds: rds}
}

func (s redisStore) Claim(ctx context.Context, key string, leaseTTL time.Duration) (idempotency.ClaimState, error) {
	return idempotency.Claim(ctx, s.rds, key, leaseTTL)
}

func (s redisStore) MarkDone(ctx context.Context, key string, successTTL time.Duration) error {
	return idempotency.MarkDone(ctx, s.rds, key, successTTL)
}

func (s redisStore) Release(ctx context.Context, key string) error {
	return idempotency.Release(ctx, s.rds, key)
}

func (s redisStore) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	n, err := s.rds.IncrCtx(ctx, key)
	if err != nil {
		return 0, err
	}
	if err := s.rds.ExpireCtx(ctx, key, int(ttl.Seconds())); err != nil {
		return n, err
	}
	return n, nil
}

// Handler 是业务方注入的回调接口。业务幂等容忍（如库存"已扣减"、券"已使用"
// 类幂等响应）由 Handle 内部按成功处理（返回 nil），骨架不感知。
type Handler[T any] interface {
	// Decode 反序列化消息体；返回 error 即按毒消息处理（一次即弃）。
	Decode(body []byte) (T, error)
	// IdempotencyKey 返回消息业务标识（如 orderId），用于幂等与重试计数 key。
	IdempotencyKey(msg T) string
	// Handle 执行全部副作用；返回 nil 表示全部成功（随后 Ack），返回 error 进入有界重试。
	Handle(ctx context.Context, msg T) error
}

// Config 骨架配置；零值字段回落默认值。
type Config struct {
	// QueueName 消费的队列名（同时用于日志与重试计数 key）
	QueueName string
	// IdempotencyService 幂等 key 的 service 段（如 "order"）
	IdempotencyService string
	// IdempotencyQueue 幂等 key 的 queue 段（如 "notify"）
	IdempotencyQueue string
	// RetryLimit 失败重投上限（默认 3），达到上限弃单
	RetryLimit int
	// BackoffBase 监督重启退避基值（默认 1s，指数退避上限 30s）
	BackoffBase time.Duration
	// ShutdownTimeout 优雅停机等待当前消息完成的上限（默认 10s）
	ShutdownTimeout time.Duration
	// IdempotencyTTL 幂等 PROCESSING 租约 TTL（默认 1h；过期即 stale 检测，可重新占坑）
	IdempotencyTTL time.Duration
	// SuccessTTL 幂等 SUCCESS 标记 TTL（默认 24h），期间重复消息直接跳过
	SuccessTTL time.Duration
}

func (c Config) withDefaults() Config {
	if c.RetryLimit <= 0 {
		c.RetryLimit = defaultRetryLimit
	}
	if c.BackoffBase <= 0 {
		c.BackoffBase = defaultBackoffBase
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = defaultShutdownTimeout
	}
	if c.IdempotencyTTL <= 0 {
		c.IdempotencyTTL = defaultIdempotencyTTL
	}
	if c.SuccessTTL <= 0 {
		c.SuccessTTL = defaultSuccessTTL
	}
	return c
}

// Acknowledger 抽象投递确认原语（*amqp.Delivery 天然满足），便于单测注入 fake。
type Acknowledger interface {
	Ack(multiple bool) error
	Reject(requeue bool) error
}

// ConnectionFactory 用于共享连接死亡后按需重拨（骨架不在构造期 Dial）。
type ConnectionFactory func() (*amqp.Connection, error)

// Consumer 是绑定具体 Handler 的 consumer 骨架。连接由调用方传入并复用；
// 若连接死亡且提供了 Factory，监督循环会重拨新连接并接管。
type Consumer[T any] struct {
	conn    *amqp.Connection
	factory ConnectionFactory
	store   Store
	handler Handler[T]
	cfg     Config
}

func New[T any](conn *amqp.Connection, factory ConnectionFactory, store Store, handler Handler[T], cfg Config) *Consumer[T] {
	return &Consumer[T]{
		conn:    conn,
		factory: factory,
		store:   store,
		handler: handler,
		cfg:     cfg.withDefaults(),
	}
}

// Run 是监督循环：consumeOnce 异常退出后按指数退避重启（上限 30s）；
// ctx 取消后等待当前消息完成，超过 ShutdownTimeout 强制关闭连接退出。
func (c *Consumer[T]) Run(ctx context.Context) {
	backoff := c.cfg.BackoffBase
	for {
		err := c.runOnceWithGracefulDeadline(ctx)
		if ctx.Err() != nil {
			logx.Infow("consumer stopped by context",
				logx.Field("queue", c.cfg.QueueName), logx.Field("err", err))
			return
		}
		// 恢复级别日志（与需人工介入的 error 级弃单/毒消息区分）
		logx.Infow("consumer exited, restarting with backoff",
			logx.Field("queue", c.cfg.QueueName), logx.Field("err", err), logx.Field("backoff", backoff))
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// runOnceWithGracefulDeadline 在独立 goroutine 中跑一轮消费，
// ctx 取消后限时等待其自然结束，超时强制关闭连接以打断 deliveries range。
func (c *Consumer[T]) runOnceWithGracefulDeadline(ctx context.Context) error {
	type result struct{ err error }
	res := make(chan result, 1)
	go func() { res <- result{c.consumeOnce(ctx)} }()

	select {
	case r := <-res:
		return r.err
	case <-ctx.Done():
		select {
		case r := <-res:
			return r.err
		case <-time.After(c.cfg.ShutdownTimeout):
			logx.Errorw("shutdown timeout exceeded, force closing amqp connection",
				logx.Field("queue", c.cfg.QueueName))
			if c.conn != nil {
				_ = c.conn.Close()
			}
			select {
			case r := <-res:
				return r.err
			case <-time.After(c.cfg.ShutdownTimeout):
				return errors.New("consumer did not return after forced connection close")
			}
		}
	}
}

// ensureConn 返回可用连接：复用现有连接；连接已死且有 Factory 时重拨接管。
func (c *Consumer[T]) ensureConn() (*amqp.Connection, error) {
	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn, nil
	}
	if c.factory == nil {
		if c.conn != nil {
			return nil, errors.New("amqp connection closed and no factory provided")
		}
		return nil, errors.New("no amqp connection provided")
	}
	conn, err := c.factory()
	if err != nil {
		return nil, err
	}
	logx.Infow("consumer redialed amqp connection", logx.Field("queue", c.cfg.QueueName))
	c.conn = conn
	return conn, nil
}

// consumeOnce 打开通道、QoS(1)、autoAck=false 消费直到通道关闭或 ctx 取消。
// 通道/连接任何异常关闭都会使 deliveries 关闭，从而统一收敛返回。
func (c *Consumer[T]) consumeOnce(ctx context.Context) error {
	conn, err := c.ensureConn()
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.Consume(
		c.cfg.QueueName,
		"",    // 消费者标签
		false, // 手动确认（spec：消息处理完成前不得确认）
		false, // 排他性
		false, // 本地消息
		false, // 等待确认
		nil,
	)
	if err != nil {
		return err
	}
	logx.Infow("consumer consuming", logx.Field("queue", c.cfg.QueueName))

	for {
		var (
			d  amqp.Delivery
			ok bool
		)
		select {
		case <-ctx.Done():
			// 优雅停机：当前无在途消息（prefetch=1 且同步处理），直接结束
			return nil
		case d, ok = <-deliveries:
			if !ok {
				// 通道/连接已关闭：正常退出本轮，由监督循环决定重启；
				// 若忽略 ok，closed channel 会持续吐出零值 Delivery 造成热循环
				return errors.New("deliveries channel closed")
			}
		}
		c.Process(ctx, d, d.Body)
		if ctx.Err() != nil {
			// 当前消息已处理完，不再取新消息
			return nil
		}
	}
}

// Process 处理单条消息（导出供测试与复用）：解码 -> 两态 claim -> 业务 -> MarkDone/Ack 或有界重试。
func (c *Consumer[T]) Process(ctx context.Context, d Acknowledger, body []byte) {
	msg, err := c.handler.Decode(body)
	if err != nil {
		// 毒消息：一次即弃 + 需人工介入级别日志
		logx.Errorw("poison message discarded",
			logx.Field("queue", c.cfg.QueueName),
			logx.Field("body", snippet(body)),
			logx.Field("err", err))
		_ = d.Reject(false)
		return
	}

	idemKey := idempotency.BuildKey(c.cfg.IdempotencyService, c.cfg.IdempotencyQueue, c.handler.IdempotencyKey(msg))
	state, err := c.store.Claim(ctx, idemKey, c.cfg.IdempotencyTTL)
	if err != nil {
		// claim 失败（结果未知）：按可重试失败处理，重投前释放可能的占坑
		c.handleFailure(ctx, d, idemKey, c.handler.IdempotencyKey(msg), err)
		return
	}
	switch state {
	case idempotency.ClaimProcessing:
		// 在途（含租约未过期的历史单态值）：跳过避免阻塞队头；丢失窗口由租约界定，
		// 过期后重放由业务幂等收敛（specs 两态键 Scenario）
		logx.Infow("message claim in progress, skipping",
			logx.Field("queue", c.cfg.QueueName), logx.Field("key", idemKey))
		_ = d.Ack(false)
		return
	case idempotency.ClaimDone:
		logx.Infow("message already succeeded, skipping",
			logx.Field("queue", c.cfg.QueueName), logx.Field("key", idemKey))
		_ = d.Ack(false)
		return
	}

	if err := c.safeHandle(ctx, msg); err != nil {
		c.handleFailure(ctx, d, idemKey, c.handler.IdempotencyKey(msg), err)
		return
	}
	// 业务全部成功：标记 SUCCESS；标记失败按业务失败处理（重投后业务幂等收敛）
	if err := c.store.MarkDone(ctx, idemKey, c.cfg.SuccessTTL); err != nil {
		logx.Errorw("failed to mark idempotency done",
			logx.Field("queue", c.cfg.QueueName), logx.Field("key", idemKey), logx.Field("err", err))
		c.handleFailure(ctx, d, idemKey, c.handler.IdempotencyKey(msg), err)
		return
	}
	_ = d.Ack(false)
}

// safeHandle 在业务回调外层兜底 panic：按一次业务失败处理，不外泄 goroutine。
func (c *Consumer[T]) safeHandle(ctx context.Context, msg T) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			logx.Errorw("handler panicked",
				logx.Field("queue", c.cfg.QueueName),
				logx.Field("panic", recovered),
				logx.Field("idempotency_key", c.handler.IdempotencyKey(msg)))
			err = errors.New("handler panicked")
		}
	}()
	return c.handler.Handle(ctx, msg)
}

// handleFailure 有界重试：未超限 -> 释放幂等标记 + Reject(true) 重投（恢复级别日志）；
// 达到上限 -> Reject(false) 弃单 + 需人工介入级别日志（含订单标识、原因、次数）。
func (c *Consumer[T]) handleFailure(ctx context.Context, d Acknowledger, idemKey, identity string, cause error) {
	count, cerr := c.store.IncrWithTTL(ctx, retryKey(c.cfg.IdempotencyQueue, identity), c.cfg.IdempotencyTTL)
	if cerr != nil {
		// 计数失败按首次失败处理（仍重投），并记录计数器异常
		logx.Errorw("failed to increment retry counter",
			logx.Field("queue", c.cfg.QueueName), logx.Field("key", retryKey(c.cfg.IdempotencyQueue, identity)), logx.Field("err", cerr))
		count = 1
	}

	if count >= int64(c.cfg.RetryLimit) {
		logx.Errorw("message abandoned after retry limit",
			logx.Field("queue", c.cfg.QueueName),
			logx.Field("identity", identity),
			logx.Field("cause", cause),
			logx.Field("attempts", count))
		_ = d.Reject(false)
		return
	}

	// 重投前释放幂等标记，保证重投消息可被正常处理（spec：失败消息重投可正常处理）
	if rerr := c.store.Release(ctx, idemKey); rerr != nil {
		logx.Errorw("failed to release idempotency key before requeue",
			logx.Field("queue", c.cfg.QueueName), logx.Field("key", idemKey), logx.Field("err", rerr))
	}
	logx.Infow("message requeued for retry",
		logx.Field("queue", c.cfg.QueueName),
		logx.Field("identity", identity),
		logx.Field("cause", cause),
		logx.Field("attempt", count))
	_ = d.Reject(true)
}

func retryKey(queue, identity string) string {
	return retryKeyPrefix + ":" + queue + ":" + identity
}

func snippet(body []byte) string {
	const maxSnippet = 256
	if len(body) <= maxSnippet {
		return string(body)
	}
	return string(body[:maxSnippet]) + "...(truncated)"
}
