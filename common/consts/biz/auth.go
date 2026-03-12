package biz

import (
	"time"
)

type CtxKey string

const (
	AuthsRpcPort        = 10000
	UserIDKey    CtxKey = "user_id"
	ClientIPKey  CtxKey = "client_ip"

	// 旧版令牌（已废弃）
	TokenExpire        = time.Hour * 2
	TokenRenewalExpire = time.Hour * 24 * 7

	// 长短令牌制常量
	LongTokenExpire  = time.Hour * 24 * 30 // 长令牌有效期：30天
	ShortTokenExpire = time.Hour * 24      // 短令牌有效期：1天

	// Redis key 前缀
	SessionKeyPrefix   = "user:session:"   // Session 存储 key 前缀
	BlacklistKeyPrefix = "user:blacklist:" // 强制下线黑名单 key 前缀

	TokenKey        = "Short-Token"
	LongTokenKey    = "Long-Token"
	RefreshTokenKey = "Long-Token"

	// 签名密钥（生产环境应从配置中心读取）
	TokenSignSecret = "go-mall-secret-key"
)
