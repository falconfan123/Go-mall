package idempotency

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// KeyPrefix is the prefix for all idempotency keys
	KeyPrefix = "idempotency"

	// 两态键状态值（design 决策 4）：PROCESSING 租约 → SUCCESS 长标记
	StateProcessing = "PROCESSING"
	StateSuccess    = "SUCCESS"

	// legacySingleStateValue 第 1 版单态 CheckAndSet 写入的值；语义上视为
	// Processing（保守正确：不能映射为 Done，否则崩溃消息会被误跳过）
	legacySingleStateValue = "1"

	// luaScript is the Lua script for atomic SETNX with TTL
	// Returns "OK" if key was set (new message), existing value if key exists (duplicate)
	luaScript = `
		if redis.call('SETNX', KEYS[1], ARGV[1]) == 1 then
			redis.call('EXPIRE', KEYS[1], ARGV[2])
			return 'OK'
		else
			return redis.call('GET', KEYS[1])
		end
	`

	// claimLua 两态占坑：key 不存在则写入 PROCESSING（PX 租约毫秒）并返回 NEW；
	// 已存在则返回当前值（PROCESSING/SUCCESS/历史单态 "1"）。
	claimLua = `
		local v = redis.call('GET', KEYS[1])
		if not v then
			redis.call('SET', KEYS[1], '` + StateProcessing + `', 'PX', ARGV[1])
			return 'NEW'
		end
		return v
	`

	// markDoneLua 成功标记：无条件覆盖为 SUCCESS（PX 毫秒）。仅应在业务全部
	// 成功后调用；失败路径走 Release（DEL）。
	markDoneLua = `
		redis.call('SET', KEYS[1], '` + StateSuccess + `', 'PX', ARGV[1])
		return 'OK'
	`
)

// ClaimState 幂等占坑结果
type ClaimState int

const (
	// ClaimNew 新消息，占坑成功（key=PROCESSING，租约=leaseTTL）
	ClaimNew ClaimState = iota
	// ClaimProcessing 已有 PROCESSING 且租约未过期（含历史单态值 "1"）——跳过等待
	ClaimProcessing
	// ClaimDone 已有 SUCCESS 标记——重复消息直接跳过
	ClaimDone
)

// CheckAndSet checks if a message has been processed and sets the key if not.
// Returns true if the message was already processed (duplicate), false if new message.
// Uses Redis Lua script for atomic SETNX + EXPIRE operation.
func CheckAndSet(ctx context.Context, rds *redis.Redis, key string, ttl time.Duration) (bool, error) {
	fullKey := fmt.Sprintf("%s:%s", KeyPrefix, key)
	ttlSeconds := int(ttl.Seconds())

	result, err := rds.EvalCtx(ctx, luaScript, []string{fullKey}, []any{"1", ttlSeconds})
	if err != nil {
		logx.Errorw("failed to check idempotency", logx.Field("key", fullKey), logx.Field("err", err))
		return false, err
	}

	if result == "OK" {
		// Key set successfully, this is a new message
		logx.Infow("idempotency key set", logx.Field("key", fullKey), logx.Field("ttl", ttl))
		return false, nil
	}

	// Key already exists, message was already processed (duplicate)
	logx.Infow("message already processed (duplicate)", logx.Field("key", fullKey), logx.Field("result", result))
	return true, nil
}

// BuildKey builds an idempotency key with the given components.
// Format: idempotency:{service}:{queue}:{message_id}
func BuildKey(service, queue, messageID string) string {
	return fmt.Sprintf("%s:%s:%s", KeyPrefix, service+":"+queue, messageID)
}

// Claim 两态占坑（design 决策 4）：key 不存在时原子写入 PROCESSING（PX=leaseTTL）
// 并返回 ClaimNew；已存在时返回当前状态（租约未过期 PROCESSING / SUCCESS / 历史
// 单态 "1" 均映射为对应 ClaimState）。租约过期后 key 自动消失，下次 Claim 即为
// ClaimNew——PX 过期即 stale 检测，无需额外时间戳比对。
func Claim(ctx context.Context, rds *redis.Redis, key string, leaseTTL time.Duration) (ClaimState, error) {
	fullKey := fmt.Sprintf("%s:%s", KeyPrefix, key)
	result, err := rds.EvalCtx(ctx, claimLua, []string{fullKey}, []any{leaseTTL.Milliseconds()})
	if err != nil {
		logx.Errorw("failed to claim idempotency", logx.Field("key", fullKey), logx.Field("err", err))
		return ClaimNew, err
	}
	switch result {
	case "NEW":
		logx.Infow("idempotency claimed (new)", logx.Field("key", fullKey), logx.Field("lease_ttl", leaseTTL))
		return ClaimNew, nil
	case StateSuccess:
		logx.Infow("idempotency already succeeded, skipping", logx.Field("key", fullKey))
		return ClaimDone, nil
	case StateProcessing, legacySingleStateValue:
		logx.Infow("idempotency claim in progress, skipping", logx.Field("key", fullKey))
		return ClaimProcessing, nil
	default:
		// 未知历史值按 Processing 保守处理（跳过、等租约过期），不误判为新消息
		logx.Infow("idempotency legacy value treated as processing", logx.Field("key", fullKey), logx.Field("value", result))
		return ClaimProcessing, nil
	}
}

// MarkDone 将幂等标记覆盖为 SUCCESS（PX=successTTL，默认建议 24h）。
// MUST 仅在全部业务副作用成功后调用；失败路径走 Release（DEL）。
func MarkDone(ctx context.Context, rds *redis.Redis, key string, successTTL time.Duration) error {
	fullKey := fmt.Sprintf("%s:%s", KeyPrefix, key)
	_, err := rds.EvalCtx(ctx, markDoneLua, []string{fullKey}, []any{successTTL.Milliseconds()})
	if err != nil {
		logx.Errorw("failed to mark idempotency done", logx.Field("key", fullKey), logx.Field("err", err))
		return err
	}
	logx.Infow("idempotency marked done", logx.Field("key", fullKey), logx.Field("success_ttl", successTTL))
	return nil
}

// Release deletes the idempotency claim key so a redelivered message can be
// processed again. Use it when message processing failed and the message will
// be requeued; the key passed in MUST be the same one given to CheckAndSet
// (i.e. built with BuildKey). Deleting a non-existent key is not an error.
func Release(ctx context.Context, rds *redis.Redis, key string) error {
	fullKey := fmt.Sprintf("%s:%s", KeyPrefix, key)
	if _, err := rds.DelCtx(ctx, fullKey); err != nil {
		logx.Errorw("failed to release idempotency key", logx.Field("key", fullKey), logx.Field("err", err))
		return err
	}
	// Recoverable-level log: releasing a claim is part of the bounded-retry path,
	// distinct from human-intervention events (abandon/poison) logged at error level.
	logx.Infow("idempotency key released", logx.Field("key", fullKey))
	return nil
}

// DefaultTTL returns a default TTL based on the delay time.
// Typically set to 2x the message delay to ensure enough time for processing.
func DefaultTTL(delay time.Duration) time.Duration {
	return delay * 2
}
