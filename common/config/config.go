package config

import (
	"fmt"
	"net/url"
	"strings"
)

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}

type RabbitMQConfig struct {
	Host  string
	Port  int
	User  string
	Pass  string
	VHost string
}

// ConsumerConfig 承载结算链路 MQ consumer 的业务参数。
// 零值安全：通过 Effective() 读取，未配置时回落到默认值。
type ConsumerConfig struct {
	// RetryLimit 业务失败重投上限，达到上限后弃单告警（默认 3）
	RetryLimit int
	// BackoffBaseMs 监督循环重启退避基值（毫秒，默认 1000，指数退避上限 30s）
	BackoffBaseMs int
	// ShutdownTimeoutMs 优雅停机时等待当前消息处理完的上限（毫秒，默认 10000）
	ShutdownTimeoutMs int
	// IdempotencyTtlSeconds 幂等 claim key（PROCESSING 状态）租约 TTL（秒，默认 3600=1h）
	IdempotencyTtlSeconds int
	// SuccessTtlSeconds 幂等 SUCCESS 标记 TTL（秒，默认 86400=24h），期间重复消息直接跳过
	SuccessTtlSeconds int
}

const (
	DefaultRetryLimit           = 3
	DefaultBackoffBaseMs        = 1000
	DefaultShutdownTimeoutMs    = 10000
	DefaultIdempotencyTtlSecond = 3600
	DefaultSuccessTtlSeconds    = 86400
)

func (c ConsumerConfig) Effective() ConsumerConfig {
	if c.RetryLimit <= 0 {
		c.RetryLimit = DefaultRetryLimit
	}
	if c.BackoffBaseMs <= 0 {
		c.BackoffBaseMs = DefaultBackoffBaseMs
	}
	if c.ShutdownTimeoutMs <= 0 {
		c.ShutdownTimeoutMs = DefaultShutdownTimeoutMs
	}
	if c.IdempotencyTtlSeconds <= 0 {
		c.IdempotencyTtlSeconds = DefaultIdempotencyTtlSecond
	}
	if c.SuccessTtlSeconds <= 0 {
		c.SuccessTtlSeconds = DefaultSuccessTtlSeconds
	}
	return c
}

type ElasticSearchConfig struct {
	Addr      string
	IndexName string
}
type GorseConfig struct {
	GorseAddr   string
	GorseApikey string
}

func (r *RabbitMQConfig) Dns() string {
	vhost := strings.TrimSpace(r.VHost)
	if vhost == "" {
		vhost = "/"
	}

	escapedVHost := "%2F"
	if vhost != "/" {
		escapedVHost = url.PathEscape(strings.TrimPrefix(vhost, "/"))
	}

	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		r.User,
		r.Pass,
		r.Host,
		r.Port,
		escapedVHost,
	)
}
