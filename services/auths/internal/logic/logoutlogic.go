package logic

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/utils/token"
	"github.com/falconfan123/Go-mall/services/auths/internal/svc"
	"github.com/falconfan123/Go-mall/services/auths/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Logout 登出（删除Session）
func (l *LogoutLogic) Logout(in *pb.LogoutReq) (*pb.LogoutRes, error) {
	res := new(pb.LogoutRes)

	// 1. 验证长令牌获取 SessionID
	longToken := in.GetLongToken()
	if longToken == "" {
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "长令牌不能为空"
		return res, nil
	}

	// 验证并解析长令牌
	sessionID, err := token.VerifyLongToken(longToken, biz.TokenSignSecret)
	if err != nil {
		l.Logger.Infow("long token verify failed during logout", logx.Field("err", err))
		res.StatusCode = code.TokenInvalid
		res.StatusMsg = "令牌无效，请重新登录"
		return res, nil
	}

	// 2. 删除 Redis 中的 Session 数据
	sessionKey := fmt.Sprintf("%s%s", biz.SessionKeyPrefix, sessionID)
	_, err = l.svcCtx.Redis.Del(sessionKey)
	if err != nil {
		l.Logger.Errorw("delete session from redis failed", logx.Field("err", err))
		res.StatusCode = code.ServerError
		res.StatusMsg = "服务器错误"
		return res, nil
	}

	l.Logger.Infow("logout successfully, session deleted",
		logx.Field("session_id", sessionID))

	res.StatusCode = 0
	res.StatusMsg = "success"

	return res, nil
}
