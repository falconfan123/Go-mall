package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/utils/cryptx"
	"github.com/falconfan123/Go-mall/common/utils/token"
	"github.com/falconfan123/Go-mall/dal/model/user"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/userspb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录方法
func (l *LoginLogic) Login(in *userspb.LoginRequest) (*userspb.LoginResponse, error) {
	// 1. Check user existence
	u, err := l.svcCtx.UsersModel.FindOneByEmail(l.ctx, sql.NullString{String: in.Email, Valid: true})
	if err != nil {
		if err == user.ErrNotFound {
			return &userspb.LoginResponse{
				StatusCode: uint32(code.UserNotExistError),
				StatusMsg:  code.UserNotExistErrorMsg,
			}, nil
		}
		return &userspb.LoginResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	// 2. Verify password
	if !cryptx.PasswordVerify(in.Password, u.PasswordHash.String) {
		return &userspb.LoginResponse{
			StatusCode: uint32(code.LoginError),
			StatusMsg:  code.LoginErrorMsg,
		}, nil
	}

	// 3. Generate tokens
	accessToken, err := token.GenerateJWT(
		uint32(u.UserId),
		"", // role
		in.Ip,
		time.Duration(l.svcCtx.Config.AuthConfig.AccessExpire)*time.Second,
	)
	if err != nil {
		l.Logger.Errorw("generate access token failed", logx.Field("err", err))
		return &userspb.LoginResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	refreshToken, err := token.GenerateJWT(
		uint32(u.UserId),
		"",
		in.Ip,
		time.Duration(l.svcCtx.Config.AuthConfig.AccessExpire*2)*time.Second,
	)
	if err != nil {
		l.Logger.Errorw("generate refresh token failed", logx.Field("err", err))
	}

	return &userspb.LoginResponse{
		StatusCode:   0,
		StatusMsg:    "success",
		UserId:       uint32(u.UserId),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
