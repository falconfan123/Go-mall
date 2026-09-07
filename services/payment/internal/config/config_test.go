package config

import (
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

// TestMustLoadPaymentYaml 校验 payment.yaml 新增配置可加载：
// Dtm（saga 发起方）与 Consumer.SuccessTtlSeconds（两态幂等键 SUCCESS TTL）。
func TestMustLoadPaymentYaml(t *testing.T) {
	var c Config
	if err := conf.LoadConfig("../../etc/payment.yaml", &c); err != nil {
		t.Fatalf("load payment.yaml: %v", err)
	}
	if c.Dtm.Server == "" || c.Dtm.BusiHost == "" {
		t.Fatalf("dtm config missing: %+v", c.Dtm)
	}
	if c.Dtm.OrderPort == 0 || c.Dtm.InventoryPort == 0 || c.Dtm.CouponsPort == 0 {
		t.Fatalf("dtm branch ports missing: %+v", c.Dtm)
	}
	if got := c.Consumer.Effective(); got.SuccessTtlSeconds != 86400 {
		t.Fatalf("SuccessTtlSeconds = %d, want 86400", got.SuccessTtlSeconds)
	}
}

// TestFallbackConfigEffective 兜底配置零值回落默认 + 显式配置保留。
func TestFallbackConfigEffective(t *testing.T) {
	got := FallbackConfig{}.Effective()
	if got.RelayTickMs != 10000 || got.ScannerTickMs != 300000 ||
		got.Scan1ThresholdMinutes != 10 || got.Scan2ThresholdMinutes != 35 ||
		got.RelayMaxAttempts != 5 || got.TombstoneMaxRetry != 3 || got.ReconHour != 2 || got.ReconDisabled {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	invalid := FallbackConfig{RelayTickMs: -1, ScannerTickMs: -1, Scan1ThresholdMinutes: -1,
		Scan2ThresholdMinutes: -1, RelayMaxAttempts: -1, TombstoneMaxRetry: -1, ReconHour: -1}.Effective()
	if invalid.RelayTickMs != 10000 || invalid.ScannerTickMs != 300000 || invalid.Scan1ThresholdMinutes != 10 ||
		invalid.Scan2ThresholdMinutes != 35 || invalid.RelayMaxAttempts != 5 || invalid.TombstoneMaxRetry != 3 ||
		invalid.ReconHour != 2 || invalid.ReconDisabled {
		t.Fatalf("invalid values should fall back: %+v", invalid)
	}
}

// TestMustLoadFallbackYaml 校验 payment.yaml 的 Fallback/Dtm.HttpAddr 可加载。
func TestMustLoadFallbackYaml(t *testing.T) {
	var c Config
	if err := conf.LoadConfig("../../etc/payment.yaml", &c); err != nil {
		t.Fatalf("load: %v", err)
	}
	if c.Dtm.HttpAddr == "" {
		t.Fatalf("Dtm.HttpAddr missing")
	}
	f := c.Fallback.Effective()
	if f.RelayTickMs != 10000 || f.Scan1ThresholdMinutes != 10 || f.ReconDisabled {
		t.Fatalf("fallback yaml values wrong: %+v", f)
	}
}
