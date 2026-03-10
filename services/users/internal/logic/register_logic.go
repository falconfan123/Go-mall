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

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册方法
func (l *RegisterLogic) Register(in *userspb.RegisterRequest) (*userspb.RegisterResponse, error) {
	if in.Password != in.ConfirmPassword {
		return &userspb.RegisterResponse{
			StatusCode: uint32(code.RePasswordError),
			StatusMsg:  code.RePasswordErrorMsg,
		}, nil
	}

	// 1. Check if email exists
	exist, err := l.svcCtx.UsersModel.FindOneByEmail(l.ctx, sql.NullString{String: in.Email, Valid: true})
	if err != nil && err != user.ErrNotFound {
		l.Logger.Errorw("check email exist failed", logx.Field("err", err))
		return &userspb.RegisterResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}
	if exist != nil {
		l.Logger.Infof("email already exists: %s", in.Email)
		return &userspb.RegisterResponse{
			StatusCode: uint32(code.UserExistError),
			StatusMsg:  code.UserExistErrorMsg,
		}, nil
	}

	// 2. Hash password
	hashedPassword := cryptx.PasswordEncrypt(in.Password)
	// 3. Create user
	newUser := &user.Users{
		Email:        sql.NullString{String: in.Email, Valid: true},
		PasswordHash: sql.NullString{String: hashedPassword, Valid: true},
		Username:     sql.NullString{String: in.Email, Valid: true}, // Default username
	}
	res, err := l.svcCtx.UsersModel.Insert(l.ctx, newUser)
	if err != nil {
		l.Logger.Errorw("create user failed", logx.Field("err", err))
		return &userspb.RegisterResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	userId, _ := res.LastInsertId()

	// 4. Generate tokens
	accessToken, err := token.GenerateJWT(
		uint32(userId),
		"", // role/username
		in.Ip,
		time.Duration(l.svcCtx.Config.AuthConfig.AccessExpire)*time.Second,
	)
	if err != nil {
		l.Logger.Errorw("generate access token failed", logx.Field("err", err))
		return &userspb.RegisterResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	refreshToken, err := token.GenerateJWT(
		uint32(userId),
		"",
		in.Ip,
		time.Duration(l.svcCtx.Config.AuthConfig.AccessExpire*2)*time.Second,
	)
	if err != nil {
		l.Logger.Errorw("generate refresh token failed", logx.Field("err", err))
		// Log error but continue as access token is generated
	}

	return &userspb.RegisterResponse{
		StatusCode:   0,
		StatusMsg:    "success",
		UserId:       uint32(userId),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
