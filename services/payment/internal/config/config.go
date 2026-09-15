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
	Consumer       config.ConsumerConfig
	Dtm            DtmConfig
	Fallback       FallbackConfig
}

// FallbackConfig 结算兜底三层（relay/scanner/对账）参数；零值回落默认。
type FallbackConfig struct {
	// RelayTickMs 发件箱接力循环 tick（毫秒，默认 10000）
	RelayTickMs int
	// ScannerTickMs 对账扫描循环 tick（毫秒，默认 300000=5min）
	ScannerTickMs int
	// Scan1ThresholdMinutes scan1 时间差阈值：paid_at 早于该分钟数才扫（默认 10）
	Scan1ThresholdMinutes int
	// Scan2ThresholdMinutes scan2 时间差阈值：created_at 早于该分钟数才关单（默认 35，> 30min delay）
	Scan2ThresholdMinutes int
	// RelayMaxAttempts outbox 重试上限，超限置 dead（默认 5）
	RelayMaxAttempts int
	// TombstoneMaxRetry 墓碑单递增 gid 重试上限（默认 3）
	TombstoneMaxRetry int
	// ReconHour 对账运行小时（本地时区，默认 2）
	ReconHour int
	// ReconDisabled 是否禁用 Stripe 日终对账（默认 false=启用；bool 零值即默认启用）
	ReconDisabled bool
}

const (
	DefaultRelayTickMs       = 10000
	DefaultScannerTickMs     = 300000
	DefaultScan1ThresholdMin = 10
	DefaultScan2ThresholdMin = 35
	DefaultRelayMaxAttempts  = 5
	DefaultTombstoneMaxRetry = 3
	DefaultReconHour         = 2
)

func (c FallbackConfig) Effective() FallbackConfig {
	if c.RelayTickMs <= 0 {
		c.RelayTickMs = DefaultRelayTickMs
	}
	if c.ScannerTickMs <= 0 {
		c.ScannerTickMs = DefaultScannerTickMs
	}
	if c.Scan1ThresholdMinutes <= 0 {
		c.Scan1ThresholdMinutes = DefaultScan1ThresholdMin
	}
	if c.Scan2ThresholdMinutes <= 0 {
		c.Scan2ThresholdMinutes = DefaultScan2ThresholdMin
	}
	if c.RelayMaxAttempts <= 0 {
		c.RelayMaxAttempts = DefaultRelayMaxAttempts
	}
	if c.TombstoneMaxRetry <= 0 {
		c.TombstoneMaxRetry = DefaultTombstoneMaxRetry
	}
	if c.ReconHour <= 0 {
		c.ReconHour = DefaultReconHour
	}
	return c
}

// DtmConfig dtm server 直连与分支回调地址（仅 payment 作为 saga 发起方需要）
type DtmConfig struct {
	// Server dtm server gRPC 直连地址（host:port）
	Server string
	// HttpAddr dtm HTTP API（query/query 墓碑探测用，默认端口 36789）
	HttpAddr string
	// BusiHost dtm 容器/集群回调分支服务的 host 前缀（dev 下为 host.docker.internal）
	BusiHost string
	// OrderPort / InventoryPort / CouponsPort 三个分支服务的 gRPC 端口
	OrderPort     int
	InventoryPort int
	CouponsPort   int
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
