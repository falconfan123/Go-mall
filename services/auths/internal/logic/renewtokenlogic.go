package logic

import (
	"context"
	"encoding/json"
	"fmt"

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

// RenewToken 续期身份
func (l *RenewTokenLogic) RenewToken(in *auths.AuthRenewalReq) (*auths.AuthRenewalRes, error) {
	res := new(auths.AuthRenewalRes)

	// 1. 验证 Long Token (Refresh Token) 签名
	sessionID, err := token.VerifySessionID(in.RefreshToken, "go-mall")
	if err != nil {
		res.StatusCode = code.TokenValid
		res.StatusMsg = code.TokenInvalidMsg
		l.Logger.Infow("refresh token verify failed", logx.Field("err", err), logx.Field("token", in.RefreshToken))
		return res, nil
	}

	// 2. 检查 Redis Session 是否存在
	sessionKey := fmt.Sprintf("auth:session:%s", sessionID)
	sessionStr, err := l.svcCtx.Redis.Get(sessionKey)
	if err != nil {
		// Redis 返回错误通常意味着 key 不存在（过期）
		res.StatusCode = code.AuthExpired
		res.StatusMsg = code.AuthExpiredMsg
		l.Logger.Infow("session expired or not found", logx.Field("session_id", sessionID))
		return res, nil
	}

	// 3. 解析 Session 数据
	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(sessionStr), &sessionData); err != nil {
		l.Logger.Errorw("session data unmarshal failed", logx.Field("err", err))
		return nil, err
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

	// 5. 生成新的 Access Token (Short Token)
	userIDFloat, _ := sessionData["user_id"].(float64)
	userID := uint32(userIDFloat)
	username, _ := sessionData["username"].(string)

	accessToken, err := token.GenerateJWT(userID, username, clientIP, biz.TokenExpire)
	if err != nil {
		l.Logger.Errorw("access token generate failed", logx.Field("err", err))
		return nil, err
	}

	// 6. 延长 Session 有效期 (Rolling Session) - 保持活跃
	l.svcCtx.Redis.Expire(sessionKey, 30*24*3600)

	res.AccessToken = accessToken
	res.RefreshToken = in.RefreshToken // 返回相同的 Long Token

	l.Logger.Infow("tokens renewed successfully via Session",
		logx.Field("user_id", userID),
		logx.Field("client_ip", clientIP))
	return res, nil
}
