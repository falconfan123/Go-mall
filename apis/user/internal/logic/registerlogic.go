package logic

import (
	"context"
	"regexp"

	"github.com/zeromicro/x/errors"

	"github.com/falconfan123/Go-mall/apis/user/internal/svc"
	"github.com/falconfan123/Go-mall/apis/user/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/auths/pb"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

// RegisterLogic is the business logic for RegisterLogic operations.
type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRegisterLogic creates a new RegisterLogic instance.
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// does something.
func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	// todo: add your logic here and delete this line

	if req.Email == "" || req.Password == "" {
		return nil, errors.New(code.LoginMessageEmpty, code.LoginMessageEmptyMsg)
	}

	// 使用RFC 5322简化版正则
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		l.Logger.Infow("邮箱格式不合法")
		return nil, errors.New(code.EmailFormatError, code.EmailFormatErrorMsg)
	}

	if req.Password != req.ConfirmPassword {
		l.Logger.Infow("密码不一致")
		return nil, errors.New(code.PasswordNotMatch, code.PasswordNotMatchMsg)

	}

	userIP := l.ctx.Value(biz.ClientIPKey).(string)

	response, err := l.svcCtx.UserRPC.Register(l.ctx, &users.RegisterRequest{
		Ip:              userIP,
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	})

	if err != nil {

		l.Logger.Errorw("call rpc register failed", logx.Field("err", err))
		return nil, errors.New(code.ServerError, err.Error())
	} else if response.StatusMsg != "" {

		return nil, errors.New(int(response.StatusCode), response.StatusMsg)

	}

	clientIP := l.ctx.Value(biz.ClientIPKey).(string)

	authrespone, err := l.svcCtx.AuthsRPC.GenerateToken(l.ctx, &authsclient.AuthGenReq{
		UserId:   response.UserId,
		Username: "",
		ClientIp: clientIP,
	})
	if err != nil {
		l.Logger.Errorw("call rpc generate token failed", logx.Field("err", err))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)

	}

	resp = &types.RegisterResponse{
		ShortToken:     authrespone.ShortToken,
		LongToken:      authrespone.LongToken,
		ShortExpiresIn: authrespone.ShortExpiresIn,
		LongExpiresIn:  authrespone.LongExpiresIn,
	}

	return resp, nil
}
