package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/utils/token"
	"github.com/falconfan123/Go-mall/services/auths/internal/svc"
	auths "github.com/falconfan123/Go-mall/services/auths/pb/auths"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ValidateToken 验证长短令牌
// 优先验证短令牌，如果短令牌过期或无效，则验证长令牌
func (l *ValidateTokenLogic) ValidateToken(in *auths.AuthValidateReq) (*auths.AuthValidateRes, error) {
	res := new(auths.AuthValidateRes)
	shortToken := in.GetShortToken()
	longToken := in.GetLongToken()
	clientIP := in.GetClientIp()

	if clientIP == "" {
		res.StatusCode = code.NotWithClientIP
		res.StatusMsg = code.NotWithClientIPMsg
		return res, nil
	}

	if shortToken == "" {
		res.StatusCode = code.AuthBlank
		res.StatusMsg = code.AuthBlankMsg
		return res, nil
	}

	// 1. 优先验证短令牌
	userID, deviceID, expireTime, err := token.VerifyShortToken(shortToken, biz.TokenSignSecret)
	shortTokenExpired := errors.Is(err, token.ErrTokenExpired)
	if err != nil && !shortTokenExpired {
		l.Logger.Infow("short token validation failed without fallback", logx.Field("err", err))
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "令牌非法，请重新登录"
		return res, nil
	}

	if longToken == "" {
		if shortTokenExpired {
			res.StatusCode = code.AuthExpired
			res.StatusMsg = code.AuthExpiredMsg
			res.NeedsRefresh = true
			return res, nil
		}
		res.StatusCode = code.AuthBlank
		res.StatusMsg = "长令牌不能为空"
		return res, nil
	}

	// 2. 无论短令牌是否过期，最终都要依赖 long token + session 做服务端裁决
	sessionID, err := token.VerifyLongToken(longToken, biz.TokenSignSecret)
	if err != nil {
		l.Logger.Infow("long token validation failed", logx.Field("err", err))
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "令牌非法，请重新登录"
		return res, nil
	}

	// 3. 根据 SessionID 查询 Redis 中的 Session 数据
	sessionKey := fmt.Sprintf("%s%s", biz.SessionKeyPrefix, sessionID)
	sessionDataStr, err := l.svcCtx.Redis.Get(sessionKey)
	if err != nil || sessionDataStr == "" {
		l.Logger.Infow("session not found in redis", logx.Field("session_id", sessionID))
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "Session不存在，请重新登录"
		return res, nil
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(sessionDataStr), &sessionData); err != nil {
		l.Logger.Errorw("parse session data failed", logx.Field("err", err))
		res.StatusCode = code.ServerError
		res.StatusMsg = "服务器错误"
		return res, nil
	}

	storedIP, _ := sessionData["client_ip"].(string)
	if storedIP != "" && storedIP != clientIP {
		res.StatusCode = code.AuthExpired
		res.StatusMsg = "IP changed, please login again"
		return res, nil
	}

	sessionUserID := uint32(sessionData["user_id"].(float64))
	sessionDeviceID, _ := sessionData["device_id"].(string)

	if !shortTokenExpired {
		if sessionUserID != userID || (sessionDeviceID != "" && sessionDeviceID != deviceID) {
			res.StatusCode = code.TokenInvalid
			res.StatusMsg = "令牌非法，请重新登录"
			return res, nil
		}

		l.Logger.Infow("short token validated successfully with active session",
			logx.Field("user_id", userID),
			logx.Field("device_id", deviceID),
			logx.Field("expire_time", expireTime),
			logx.Field("session_id", sessionID))

		res.StatusCode = code.Success
		res.StatusMsg = "success"
		res.UserId = userID
		return res, nil
	}

	// 4. 短令牌过期时，长令牌和 Session 仍有效，要求客户端刷新短令牌
	if longToken == "" {
		res.StatusCode = code.AuthExpired
		res.StatusMsg = code.AuthExpiredMsg
		res.NeedsRefresh = true
		return res, nil
	}
	l.Logger.Infow("short token expired but session is still valid",
		logx.Field("user_id", sessionUserID),
		logx.Field("session_id", sessionID))

	res.StatusCode = code.AuthExpired
	res.StatusMsg = code.TokenRenewedMsg
	res.UserId = sessionUserID
	res.NeedsRefresh = true
	return res, nil
}
