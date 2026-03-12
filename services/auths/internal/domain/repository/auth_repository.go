package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrLogoutTimeNotFound = errors.New("logout time not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
)

// TokenRepository Token仓储接口
type TokenRepository interface {
	// GenerateToken 生成Token
	GenerateToken(ctx context.Context, userID int64, clientIP string, expiry time.Duration) (string, error)

	// ParseToken 解析Token
	ParseToken(ctx context.Context, tokenString string) (*TokenClaims, error)

	// ValidateToken 验证Token
	ValidateToken(ctx context.Context, tokenString string, clientIP string) (*TokenClaims, error)

	// GetLogoutTime 获取用户登出时间
	GetLogoutTime(ctx context.Context, userID int64) (*time.Time, error)

	// SetLogoutTime 设置用户登出时间
	SetLogoutTime(ctx context.Context, userID int64, logoutTime time.Time) error
}

// TokenClaims Token声明
type TokenClaims struct {
	UserID    int64     // 用户ID
	ClientIP  string    // 客户端IP
	IssuedAt  time.Time // 签发时间
	ExpiresAt time.Time // 过期时间
}
