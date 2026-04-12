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
	return fmt.Sprintf("%s:%s:%s", service, queue, messageID)
}

// DefaultTTL returns a default TTL based on the delay time.
// Typically set to 2x the message delay to ensure enough time for processing.
func DefaultTTL(delay time.Duration) time.Duration {
	return delay * 2
}
