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
