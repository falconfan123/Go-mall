package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/falconfan123/Go-mall/common/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	PostgresConfig config.PostgresConfig
	RedisConf      redis.RedisConf
	Stripe         StripeConfig
	OrderRpc       zrpc.RpcClientConf
	RabbitMQConfig config.RabbitMQConfig
	PrometheusExt  PrometheusExtConf
}

type StripeConfig struct {
	APIKey            string
	SuccessURL        string
	CancelURL         string
	WebhookSecret     string
	WebhookPort       int
	RequestTimeoutMs  int64
	MaxNetworkRetries int64
}

// ValidateStripeConfig validates the Stripe configuration
func ValidateStripeConfig(isLocalDev bool) error {
	cfg := GetStripeConfig()

	// Check if using localhost/127.0.0.1 in non-local mode
	if !isLocalDev {
		if strings.Contains(cfg.SuccessURL, "127.0.0.1") ||
			strings.Contains(cfg.SuccessURL, "localhost") ||
			strings.Contains(cfg.CancelURL, "127.0.0.1") ||
			strings.Contains(cfg.CancelURL, "localhost") {
			return fmt.Errorf("Stripe callback URLs must not use 127.0.0.1 or localhost in non-local mode")
		}

		if cfg.SuccessURL == "" || cfg.CancelURL == "" {
			return fmt.Errorf("Stripe SuccessURL and CancelURL must be configured in non-local mode")
		}

		// Validate URL format
		if !strings.HasPrefix(cfg.SuccessURL, "http://") && !strings.HasPrefix(cfg.SuccessURL, "https://") {
			return fmt.Errorf("Stripe SuccessURL must include protocol (http:// or https://)")
		}
		if !strings.HasPrefix(cfg.CancelURL, "http://") && !strings.HasPrefix(cfg.CancelURL, "https://") {
			return fmt.Errorf("Stripe CancelURL must include protocol (http:// or https://)")
		}
	}

	return nil
}

// GetStripeConfig returns the current Stripe configuration
func GetStripeConfig() StripeConfig {
	return StripeConfig{
		APIKey:            os.Getenv("STRIPE_API_KEY"),
		SuccessURL:        os.Getenv("STRIPE_SUCCESS_URL"),
		CancelURL:         os.Getenv("STRIPE_CANCEL_URL"),
		WebhookSecret:     os.Getenv("STRIPE_WEBHOOK_SECRET"),
		WebhookPort:       11112,
		RequestTimeoutMs:  8000,
		MaxNetworkRetries: 2,
	}
}

type PrometheusExtConf struct {
	Host string
	Port int
	Path string
}
