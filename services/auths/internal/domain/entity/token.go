package entity

import "time"

// TokenInfo Token信息实体
type TokenInfo struct {
	UserID    int64     // 用户ID
	ClientIP  string    // 客户端IP
	IssuedAt  time.Time // 签发时间
	ExpiresAt time.Time // 过期时间
}

// NewTokenInfo 创建Token信息
func NewTokenInfo(userID int64, clientIP string, issuedAt, expiresAt time.Time) *TokenInfo {
	return &TokenInfo{
		UserID:    userID,
		ClientIP:  clientIP,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}
}

// IsExpired 检查是否过期
func (t *TokenInfo) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// AuthResult 认证结果
type AuthResult struct {
	UserID     int64  // 用户ID
	StatusCode int64  // 状态码
	StatusMsg  string // 状态信息
	IsValid    bool   // 是否有效
}

// NewAuthResult 创建认证结果
func NewAuthResult(userID int64, statusCode int64, statusMsg string, isValid bool) *AuthResult {
	return &AuthResult{
		UserID:     userID,
		StatusCode: statusCode,
		StatusMsg:  statusMsg,
		IsValid:    isValid,
	}
}
