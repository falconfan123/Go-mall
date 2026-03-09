package middleware

import (
	"context"
	"jijizhazha1024/go-mall/common/consts/code"
	"jijizhazha1024/go-mall/common/response"
	"net/http"

	"github.com/zeromicro/go-zero/core/limit"
)

// Limiter defines the interface for rate limiters
type Limiter interface {
	Allow(ctx context.Context) (bool, error)
}

// TokenLimiterWrapper wraps a limit.TokenLimiter to implement the Limiter interface
type TokenLimiterWrapper struct {
	limiter *limit.TokenLimiter
}

func (w *TokenLimiterWrapper) Allow(ctx context.Context) (bool, error) {
	return w.limiter.AllowCtx(ctx), nil
}

// NewTokenLimiterWrapper creates a new TokenLimiterWrapper
func NewTokenLimiterWrapper(limiter *limit.TokenLimiter) Limiter {
	return &TokenLimiterWrapper{limiter: limiter}
}

// PeriodLimiterWrapper wraps a limit.PeriodLimit to implement the Limiter interface
type PeriodLimiterWrapper struct {
	limiter *limit.PeriodLimit
	key     string
}

func (w *PeriodLimiterWrapper) Allow(ctx context.Context) (bool, error) {
	code, err := w.limiter.TakeCtx(ctx, w.key)
	if err != nil {
		return false, err
	}
	return code == limit.Allowed, nil
}

// NewPeriodLimiterWrapper creates a new PeriodLimiterWrapper
func NewPeriodLimiterWrapper(limiter *limit.PeriodLimit, key string) Limiter {
	return &PeriodLimiterWrapper{limiter: limiter, key: key}
}

// RateLimiterMiddleware is a middleware that limits the rate of requests
func RateLimiterMiddleware(limiter Limiter) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			// Check if the request is allowed by the rate limiter
			if allowed, err := limiter.Allow(ctx); err != nil || !allowed {
				// If there's an error or the request is not allowed, return 429 Too Many Requests
				response.Fail(w, code.RateLimitExceeded)
				return
			}
			// If allowed, proceed to the next handler
			next(w, r)
		}
	}
}
