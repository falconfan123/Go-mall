package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/utils/token"

	"github.com/falconfan123/Go-mall/services/auths/auths"
	"github.com/falconfan123/Go-mall/services/auths/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RenewTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRenewTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RenewTokenLogic {
	return &RenewTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RenewToken 续期身份（刷新短令牌）
func (l *RenewTokenLogic) RenewToken(in *auths.AuthRenewalReq) (*auths.AuthRenewalRes, error) {
	res := new(auths.AuthRenewalRes)

	// 1. 验证 Long Token 签名
	longToken := in.GetLongToken()
	sessionID, err := token.VerifyLongToken(longToken, biz.TokenSignSecret)
	if err != nil {
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "Session已过期，请重新登录"
		l.Logger.Infow("long token verify failed", logx.Field("err", err))
		return res, nil
	}

	// 2. 检查 Redis Session 是否存在
	sessionKey := fmt.Sprintf("%s%s", biz.SessionKeyPrefix, sessionID)
	sessionStr, err := l.svcCtx.Redis.Get(sessionKey)
	if err != nil || sessionStr == "" {
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "Session不存在，请重新登录"
		l.Logger.Infow("session not found or expired", logx.Field("session_id", sessionID))
		return res, nil
	}

	// 3. 解析 Session 数据
	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(sessionStr), &sessionData); err != nil {
		l.Logger.Errorw("session data unmarshal failed", logx.Field("err", err))
		res.StatusCode = code.ServerError
		res.StatusMsg = "服务器错误"
		return res, nil
	}

	// 4. 验证 Client IP
	clientIP := in.GetClientIp()
	if clientIP == "" {
		res.StatusCode = code.NotWithClientIP
		res.StatusMsg = code.NotWithClientIPMsg
		return res, nil
	}

	storedIP, ok := sessionData["client_ip"].(string)
	if ok && storedIP != clientIP {
		res.StatusCode = code.AuthExpired
		res.StatusMsg = "IP changed, please login again"
		l.Logger.Infow("client ip changed",
			logx.Field("old_ip", storedIP),
			logx.Field("new_ip", clientIP))
		return res, nil
	}

	// 5. 获取用户信息
	userIDFloat, _ := sessionData["user_id"].(float64)
	userID := uint32(userIDFloat)
	deviceID, _ := sessionData["device_id"].(string)

	// 6. 生成新的 Short Token
	newShortToken := token.GenerateShortToken(userID, deviceID, biz.ShortTokenExpire, biz.TokenSignSecret)

	// 7. 延长 Session 有效期 (Rolling Session) - 保持活跃
	l.svcCtx.Redis.Expire(sessionKey, int(biz.LongTokenExpire.Seconds()))

	// 计算新的过期时间
	expireTime := time.Now().Add(biz.ShortTokenExpire).Unix()

	res.ShortToken = newShortToken
	res.ExpiresIn = expireTime

	l.Logger.Infow("short token refreshed successfully",
		logx.Field("user_id", userID),
		logx.Field("client_ip", clientIP),
		logx.Field("session_id", sessionID))

	res.StatusCode = 0
	res.StatusMsg = "success"

	return res, nil
}
