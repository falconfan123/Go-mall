package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/utils/token"
	"github.com/falconfan123/Go-mall/services/auths/internal/svc"
	auths "github.com/falconfan123/Go-mall/services/auths/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTokenLogic {
	return &GenerateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GenerateToken 生成长短令牌
// 长令牌：SessionID + HMAC-SHA256签名，30天有效期，存储在Redis中
// 短令牌：user_id.device_id.expire_time.signature，1天有效期，自包含验证
func (l *GenerateTokenLogic) GenerateToken(in *auths.AuthGenReq) (*auths.AuthGenRes, error) {
	res := new(auths.AuthGenRes)
	clientIP := in.GetClientIp()
	if clientIP == "" {
		res.StatusCode = code.NotWithClientIP
		res.StatusMsg = code.NotWithClientIPMsg
		l.Logger.Infow("client ip is empty", logx.Field("user_id", in.UserId))
		return res, nil
	}

	// 生成设备ID
	deviceID := in.GetDeviceId()
	if deviceID == "" {
		deviceID = token.GenerateDeviceID()
	}

	// 1. 生成 SessionID (用于长令牌)
	sessionID := token.GenerateSessionID()

	// 2. 存储 Session 到 Redis（长令牌）
	// 有效期：30天
	sessionData := map[string]interface{}{
		"user_id":    in.UserId,
		"username":   in.Username,
		"device_id":  deviceID,
		"client_ip":  clientIP,
		"login_time": time.Now().Unix(),
	}
	sessionBytes, _ := json.Marshal(sessionData)
	sessionKey := fmt.Sprintf("%s%s", biz.SessionKeyPrefix, sessionID)

	// 30 days = 30 * 24 * 3600 seconds
	err := l.svcCtx.Redis.Setex(sessionKey, string(sessionBytes), int(biz.LongTokenExpire.Seconds()))
	if err != nil {
		l.Logger.Errorw("redis set session failed", logx.Field("err", err))
		return nil, err
	}

	// 3. 生成长令牌 (Long Token) - 格式: sessionID.signature
	longToken := token.GenerateLongToken(sessionID, biz.TokenSignSecret)

	// 4. 生成短令牌 (Short Token) - 格式: user_id.device_id.expire_time.signature
	shortToken := token.GenerateShortToken(in.UserId, deviceID, biz.ShortTokenExpire, biz.TokenSignSecret)

	// 计算过期时间
	now := time.Now()
	shortExpireTime := now.Add(biz.ShortTokenExpire).Unix()
	longExpireTime := now.Add(biz.LongTokenExpire).Unix()

	// 返回长短令牌
	l.Logger.Infow("tokens generated successfully (Dual Token Strategy)",
		logx.Field("user_id", in.UserId),
		logx.Field("client_ip", clientIP),
		logx.Field("session_id", sessionID),
		logx.Field("device_id", deviceID))

	res.ShortToken = shortToken
	res.LongToken = longToken
	res.ShortExpiresIn = shortExpireTime
	res.LongExpiresIn = longExpireTime
	return res, nil
}
