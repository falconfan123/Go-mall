package config

import "testing"

func TestRabbitMQConfigDns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  RabbitMQConfig
		want string
	}{
		{
			name: "default vhost slash is encoded",
			cfg: RabbitMQConfig{
				Host:  "127.0.0.1",
				Port:  5672,
				User:  "admin",
				Pass:  "admin",
				VHost: "/",
			},
			want: "amqp://admin:admin@127.0.0.1:5672/%2F",
		},
		{
			name: "empty vhost falls back to slash",
			cfg: RabbitMQConfig{
				Host: "mq",
				Port: 5672,
				User: "u",
				Pass: "p",
			},
			want: "amqp://u:p@mq:5672/%2F",
		},
		{
			name: "custom vhost keeps path segment",
			cfg: RabbitMQConfig{
				Host:  "mq",
				Port:  5672,
				User:  "u",
				Pass:  "p",
				VHost: "orders",
			},
			want: "amqp://u:p@mq:5672/orders",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.cfg.Dns(); got != tc.want {
				t.Fatalf("Dns() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConsumerConfigEffective(t *testing.T) {
	t.Parallel()

	// 零值回落默认：重试上限 3、退避 1s、停机等待 10s、幂等租约 1h、SUCCESS 标记 24h
	got := ConsumerConfig{}.Effective()
	if got.RetryLimit != 3 || got.BackoffBaseMs != 1000 ||
		got.ShutdownTimeoutMs != 10000 || got.IdempotencyTtlSeconds != 3600 ||
		got.SuccessTtlSeconds != 86400 {
		t.Fatalf("unexpected defaults: %+v", got)
	}

	// 显式配置保留
	custom := ConsumerConfig{
		RetryLimit:            5,
		BackoffBaseMs:         500,
		ShutdownTimeoutMs:     3000,
		IdempotencyTtlSeconds: 7200,
		SuccessTtlSeconds:     3600,
	}.Effective()
	if custom != (ConsumerConfig{RetryLimit: 5, BackoffBaseMs: 500, ShutdownTimeoutMs: 3000, IdempotencyTtlSeconds: 7200, SuccessTtlSeconds: 3600}) {
		t.Fatalf("custom config altered: %+v", custom)
	}

	// 非法值回落默认
	invalid := ConsumerConfig{RetryLimit: -1, BackoffBaseMs: -1, ShutdownTimeoutMs: -1, IdempotencyTtlSeconds: -1, SuccessTtlSeconds: -1}.Effective()
	if invalid != (ConsumerConfig{RetryLimit: 3, BackoffBaseMs: 1000, ShutdownTimeoutMs: 10000, IdempotencyTtlSeconds: 3600, SuccessTtlSeconds: 86400}) {
		t.Fatalf("invalid values should fall back to defaults: %+v", invalid)
	}
}
