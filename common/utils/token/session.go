package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

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
