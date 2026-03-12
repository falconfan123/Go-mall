package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/x/errors"

	"github.com/falconfan123/Go-mall/apis/user/internal/svc"
	"github.com/falconfan123/Go-mall/apis/user/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/auths/pb"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

// LoginLogic is the business logic for LoginLogic operations.
type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic creates a new LoginLogic instance.
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// does something.
func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// todo: add your logic here and delete this line

	if req.Email == "" || req.Password == "" {
		return nil, errors.New(code.LoginMessageEmpty, code.LoginMessageEmptyMsg)
	}

	loginres, err := l.svcCtx.UserRPC.Login(l.ctx, &users.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {

		l.Logger.Errorw("call rpc login failed", logx.Field("err", err))
		fmt.Println("loginres:", loginres)
		fmt.Println("err:", err)
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	} else if loginres.StatusMsg != "" {

		return nil, errors.New(int(loginres.StatusCode), loginres.StatusMsg)

	}

	clientIP := l.ctx.Value(biz.ClientIPKey).(string)

	authrespone, err := l.svcCtx.AuthsRPC.GenerateToken(l.ctx, &authsclient.AuthGenReq{
		UserId:   loginres.UserId,
		Username: loginres.UserName,
		ClientIp: clientIP,
	})
	if err != nil {
		l.Logger.Errorw("call rpc  auth token failed", logx.Field("err", err))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)

	}

	resp = &types.LoginResponse{
		ShortToken:     authrespone.ShortToken,
		LongToken:      authrespone.LongToken,
		ShortExpiresIn: authrespone.ShortExpiresIn,
		LongExpiresIn:  authrespone.LongExpiresIn,
	}

	return resp, nil
}
