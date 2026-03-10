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

// GenerateToken 生成token
func (l *GenerateTokenLogic) GenerateToken(in *auths.AuthGenReq) (*auths.AuthGenRes, error) {
	res := new(auths.AuthGenRes)
	clientIP := in.GetClientIp()
	if clientIP == "" {
		res.StatusCode = code.NotWithClientIP
		res.StatusMsg = code.NotWithClientIPMsg
		l.Logger.Infow("client ip is empty", logx.Field("user_id", in.UserId))
		return res, nil
	}

	// 1. 生成 Access Token (Short Token) - JWT
	// 有效期较短，如 1 天 (biz.TokenExpire)
	accessToken, err := token.GenerateJWT(in.UserId, in.Username, clientIP, biz.TokenExpire)
	if err != nil {
		l.Logger.Errorw("access token generate failed",
			logx.Field("err", err),
			logx.Field("client_ip", clientIP),
			logx.Field("user_id", in.UserId))
		return nil, err
	}

	// 2. 生成 SessionID (Long Token Base)
	sessionID := token.GenerateSessionID()

	// 3. 存储 Session 到 Redis
	// 有效期较长，如 30 天
	sessionData := map[string]interface{}{
		"user_id":    in.UserId,
		"username":   in.Username,
		"client_ip":  clientIP,
		"login_time": time.Now().Unix(),
	}
	sessionBytes, _ := json.Marshal(sessionData)
	sessionKey := fmt.Sprintf("auth:session:%s", sessionID)

	// 30 days = 30 * 24 * 3600 seconds
	err = l.svcCtx.Redis.Setex(sessionKey, string(sessionBytes), 30*24*3600)
	if err != nil {
		l.Logger.Errorw("redis set session failed", logx.Field("err", err))
		return nil, err
	}

	// 4. 对 SessionID 签名生成 Long Token (Refresh Token)
	refreshToken := token.SignSessionID(sessionID, "go-mall")

	// 返回 access token 和 refresh token
	l.Logger.Infow("tokens generated successfully (Dual Token Strategy)",
		logx.Field("user_id", in.UserId),
		logx.Field("client_ip", clientIP),
		logx.Field("session_id", sessionID))

	res.AccessToken = accessToken
	res.RefreshToken = refreshToken
	return res, nil
}
