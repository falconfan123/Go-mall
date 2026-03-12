package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateSessionID 生成 SessionID (UUID)
func GenerateSessionID() string {
	return uuid.New().String()
}

// SignSessionID 对 SessionID 进行签名，生成 Long Token
// 格式: sessionID.signature
func SignSessionID(sessionID string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(sessionID))
	signature := hex.EncodeToString(h.Sum(nil))
	return sessionID + "." + signature
}

// VerifySessionID 验证 Long Token 签名，返回原始 SessionID
func VerifySessionID(signedToken string, secret string) (string, error) {
	parts := strings.Split(signedToken, ".")
	if len(parts) != 2 {
		return "", errors.New("invalid token format")
	}
	sessionID := parts[0]
	signature := parts[1]

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(sessionID))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if signature != expectedSignature {
		return "", errors.New("invalid signature")
	}
	return sessionID, nil
}

// ShortTokenData 短令牌数据结构
type ShortTokenData struct {
	UserID     uint32 `json:"user_id"`
	DeviceID   string `json:"device_id"`
	ExpireTime int64  `json:"expire_time"`
}

// GenerateShortToken 生成短令牌
// 格式: user_id.device_id.expire_time.signature
func GenerateShortToken(userID uint32, deviceID string, expireDuration time.Duration, secret string) string {
	expireTime := time.Now().Add(expireDuration).Unix()
	data := fmt.Sprintf("%d.%s.%d", userID, deviceID, expireTime)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%d.%s.%d.%s", userID, deviceID, expireTime, signature)
}

// VerifyShortToken 验证短令牌
// 返回: userID, deviceID, expireTime, error
func VerifyShortToken(shortToken string, secret string) (uint32, string, int64, error) {
	parts := strings.Split(shortToken, ".")
	if len(parts) != 4 {
		return 0, "", 0, errors.New("invalid token format")
	}

	var userID uint32
	fmt.Sscanf(parts[0], "%d", &userID)
	deviceID := parts[1]
	var expireTime int64
	fmt.Sscanf(parts[2], "%d", &expireTime)
	signature := parts[3]

	// 验证签名
	data := fmt.Sprintf("%d.%s.%d", userID, deviceID, expireTime)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if signature != expectedSignature {
		return 0, "", 0, errors.New("invalid signature")
	}

	// 验证过期时间
	if time.Now().Unix() > expireTime {
		return 0, "", 0, errors.New("token expired")
	}

	return userID, deviceID, expireTime, nil
}

// GenerateDeviceID 生成设备ID
func GenerateDeviceID() string {
	return uuid.New().String()
}

// LongTokenData 长令牌数据结构
type LongTokenData struct {
	SessionID  string `json:"session_id"`
	ExpireTime int64  `json:"expire_time"`
}

// GenerateLongToken 生成长令牌（包含 SessionID 和签名）
// 格式: sessionID.signature
func GenerateLongToken(sessionID string, secret string) string {
	return SignSessionID(sessionID, secret)
}

// VerifyLongToken 验证长令牌
// 返回: sessionID, error
func VerifyLongToken(longToken string, secret string) (string, error) {
	return VerifySessionID(longToken, secret)
}
