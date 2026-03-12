package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/utils/token"
	"github.com/falconfan123/Go-mall/services/auths/pb"
	"github.com/falconfan123/Go-mall/services/auths/internal/svc"

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
func (l *ValidateTokenLogic) ValidateToken(in *pb.AuthValidateReq) (*pb.AuthValidateRes, error) {
	res := new(pb.AuthValidateRes)
	tk := in.GetToken()

	// 1. 首先尝试验证短令牌
	if tk != "" {
		userID, deviceID, expireTime, err := token.VerifyShortToken(tk, biz.TokenSignSecret)
		if err == nil {
			// 短令牌验证成功
			l.Logger.Infow("short token validated successfully",
				logx.Field("user_id", userID),
				logx.Field("device_id", deviceID),
				logx.Field("expire_time", expireTime))

			res.StatusCode = code.Success
			res.StatusMsg = "success"
			res.UserId = userID
			return res, nil
		}

		// 短令牌验证失败，可能过期或无效
		l.Logger.Infow("short token validation failed, trying long token", logx.Field("err", err))
	}

	// 2. 短令牌验证失败，尝试验证长令牌
	// Token可能是长令牌，需要验证长令牌
	sessionID, err := token.VerifyLongToken(tk, biz.TokenSignSecret)
	if err != nil {
		l.Logger.Infow("long token validation failed", logx.Field("err", err))
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "令牌非法，请重新登录"
		return res, nil
	}

	// 3. 根据SessionID查询Redis中的Session数据
	sessionKey := fmt.Sprintf("%s%s", biz.SessionKeyPrefix, sessionID)
	sessionDataStr, err := l.svcCtx.Redis.Get(sessionKey)
	if err != nil || sessionDataStr == "" {
		l.Logger.Infow("session not found in redis", logx.Field("session_id", sessionID))
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "Session不存在，请重新登录"
		return res, nil
	}

	// 解析Session数据
	var sessionData map[string]interface{}
	err = json.Unmarshal([]byte(sessionDataStr), &sessionData)
	if err != nil {
		l.Logger.Errorw("parse session data failed", logx.Field("err", err))
		res.StatusCode = code.ServerError
		res.StatusMsg = "服务器错误"
		return res, nil
	}

	userID := uint32(sessionData["user_id"].(float64))
	deviceID := sessionData["device_id"].(string)

	// 4. 生新的短令牌（有效期重新计算为1天）
	_ = token.GenerateShortToken(userID, deviceID, biz.ShortTokenExpire, biz.TokenSignSecret)

	l.Logger.Infow("long token validated successfully",
		logx.Field("user_id", userID),
		logx.Field("session_id", sessionID))

	res.StatusCode = code.Success
	res.StatusMsg = "success"
	res.UserId = userID

	return res, nil
}
