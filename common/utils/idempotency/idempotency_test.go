package idempotency

import (
	"context"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func newTestRedis(t *testing.T) *redis.Redis {
	t.Helper()
	rds := redis.MustNewRedis(redis.RedisConf{Host: "localhost:6379", Type: "node"})
	if !rds.PingCtx(context.Background()) {
		t.Skip("local redis unreachable, skip")
	}
	return rds
}

// TestReleaseAllowsReclaim 验证 Release 语义闭环：
// CheckAndSet 占坑 -> 重复投递被识别 -> Release 释放 -> CheckAndSet 可重新占坑。
func TestReleaseAllowsReclaim(t *testing.T) {
	ctx := context.Background()
	rds := newTestRedis(t)

	key := BuildKey("order", "notify", "test-release-orders")
	fullKey := KeyPrefix + ":" + key
	t.Cleanup(func() { _, _ = rds.DelCtx(ctx, fullKey) })
	_, _ = rds.DelCtx(ctx, fullKey)

	// 首次占坑：新消息
	processed, err := CheckAndSet(ctx, rds, key, time.Minute)
	if err != nil {
		t.Fatalf("first CheckAndSet: %v", err)
	}
	if processed {
		t.Fatal("first claim should not be reported as processed")
	}

	// 重复投递：已被占坑
	processed, err = CheckAndSet(ctx, rds, key, time.Minute)
	if err != nil {
		t.Fatalf("second CheckAndSet: %v", err)
	}
	if !processed {
		t.Fatal("second claim should be reported as processed")
	}

	// 释放后重新占坑
	if err := Release(ctx, rds, key); err != nil {
		t.Fatalf("Release: %v", err)
	}
	processed, err = CheckAndSet(ctx, rds, key, time.Minute)
	if err != nil {
		t.Fatalf("third CheckAndSet: %v", err)
	}
	if processed {
		t.Fatal("claim after Release should be treated as new message")
	}
}

// TestReleaseNonExistentKeyIsNoOp 验证释放不存在的 key 不报错（幂等 DEL）。
func TestReleaseNonExistentKeyIsNoOp(t *testing.T) {
	ctx := context.Background()
	rds := newTestRedis(t)

	key := BuildKey("order", "notify", "test-release-missing")
	fullKey := KeyPrefix + ":" + key
	t.Cleanup(func() { _, _ = rds.DelCtx(ctx, fullKey) })
	_, _ = rds.DelCtx(ctx, fullKey)

	if err := Release(ctx, rds, key); err != nil {
		t.Fatalf("Release on missing key should not error: %v", err)
	}
}

// TestClaimMarkDoneLifecycle 两态键生命周期：Claim(New) → MarkDone → Claim(Done)。
func TestClaimMarkDoneLifecycle(t *testing.T) {
	ctx := context.Background()
	rds := newTestRedis(t)

	key := BuildKey("order", "delay", "test-claim-lifecycle")
	fullKey := KeyPrefix + ":" + key
	t.Cleanup(func() { _, _ = rds.DelCtx(ctx, fullKey) })
	_, _ = rds.DelCtx(ctx, fullKey)

	st, err := Claim(ctx, rds, key, time.Minute)
	if err != nil || st != ClaimNew {
		t.Fatalf("first Claim = %v, %v; want ClaimNew", st, err)
	}
	st, err = Claim(ctx, rds, key, time.Minute)
	if err != nil || st != ClaimProcessing {
		t.Fatalf("second Claim = %v, %v; want ClaimProcessing", st, err)
	}
	if err := MarkDone(ctx, rds, key, time.Minute); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	st, err = Claim(ctx, rds, key, time.Minute)
	if err != nil || st != ClaimDone {
		t.Fatalf("third Claim = %v, %v; want ClaimDone", st, err)
	}
}

// TestClaimLegacyValueTreatedAsProcessing 历史单态值 "1" 必须按 Processing 处理
// （不能映射为 Done，否则崩溃消息会被误跳过）。
func TestClaimLegacyValueTreatedAsProcessing(t *testing.T) {
	ctx := context.Background()
	rds := newTestRedis(t)

	key := BuildKey("order", "delay", "test-claim-legacy")
	fullKey := KeyPrefix + ":" + key
	t.Cleanup(func() { _, _ = rds.DelCtx(ctx, fullKey) })
	_, _ = rds.DelCtx(ctx, fullKey)

	if err := rds.SetexCtx(ctx, fullKey, "1", 60); err != nil {
		t.Fatalf("seed legacy value: %v", err)
	}
	st, err := Claim(ctx, rds, key, time.Minute)
	if err != nil || st != ClaimProcessing {
		t.Fatalf("Claim on legacy value = %v, %v; want ClaimProcessing", st, err)
	}
}

// TestClaimLeaseExpiryAllowsReclaim 租约（PX）过期即 stale 检测：过期后可重新占坑。
func TestClaimLeaseExpiryAllowsReclaim(t *testing.T) {
	ctx := context.Background()
	rds := newTestRedis(t)

	key := BuildKey("order", "delay", "test-claim-expiry")
	fullKey := KeyPrefix + ":" + key
	t.Cleanup(func() { _, _ = rds.DelCtx(ctx, fullKey) })
	_, _ = rds.DelCtx(ctx, fullKey)

	if st, err := Claim(ctx, rds, key, 2*time.Second); err != nil || st != ClaimNew {
		t.Fatalf("first Claim = %v, %v; want ClaimNew", st, err)
	}
	if st, _ := Claim(ctx, rds, key, 2*time.Second); st != ClaimProcessing {
		t.Fatalf("claim within lease should be ClaimProcessing")
	}
	time.Sleep(2500 * time.Millisecond) // 租约过期
	st, err := Claim(ctx, rds, key, 2*time.Second)
	if err != nil || st != ClaimNew {
		t.Fatalf("Claim after lease expiry = %v, %v; want ClaimNew", st, err)
	}
}

// TestReleaseResetsProcessingClaim 失败路径：Release 删除 PROCESSING 后可重新占坑。
func TestReleaseResetsProcessingClaim(t *testing.T) {
	ctx := context.Background()
	rds := newTestRedis(t)

	key := BuildKey("order", "delay", "test-claim-release")
	fullKey := KeyPrefix + ":" + key
	t.Cleanup(func() { _, _ = rds.DelCtx(ctx, fullKey) })
	_, _ = rds.DelCtx(ctx, fullKey)

	if st, _ := Claim(ctx, rds, key, time.Minute); st != ClaimNew {
		t.Fatalf("first claim should be ClaimNew")
	}
	if st, _ := Claim(ctx, rds, key, time.Minute); st != ClaimProcessing {
		t.Fatalf("duplicate claim should be ClaimProcessing")
	}
	if err := Release(ctx, rds, key); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if st, err := Claim(ctx, rds, key, time.Minute); err != nil || st != ClaimNew {
		t.Fatalf("Claim after Release = %v, %v; want ClaimNew", st, err)
	}
}
