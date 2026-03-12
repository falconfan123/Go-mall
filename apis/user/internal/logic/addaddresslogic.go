package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/apis/user/internal/svc"
	"github.com/falconfan123/Go-mall/apis/user/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type AddAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAddressLogic {
	return &AddAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddAddressLogic) AddAddress(req *types.AddAddressRequest) (resp *types.AddAddressResponse, err error) {

	//校验
	if req.City == "" || req.DetailedAddress == "" || req.PhoneNumber == "" || req.Province == "" {

		l.Logger.Errorw("用户信息为空", logx.Field("err", err))
		return nil, errors.New(code.Fail, "user informaition empty")

	}

	user_ip := l.ctx.Value(biz.ClientIPKey).(string)

	user_id := l.ctx.Value(biz.UserIDKey).(uint32)
	addaddressresp, err := l.svcCtx.UserRpc.AddAddress(l.ctx, &users.AddAddressRequest{
		Ip: user_ip,

		UserId:          user_id,
		RecipientName:   req.RecipientName,
		Province:        req.Province,
		City:            req.City,
		PhoneNumber:     req.PhoneNumber,
		DetailedAddress: req.DetailedAddress,

		IsDefault: req.IsDefault,
	})

	if err != nil {

		l.Logger.Errorw("call rpc add address add failed", logx.Field("err", err))

		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	} else if addaddressresp.StatusMsg != "" {

		return nil, errors.New(int(addaddressresp.StatusCode), addaddressresp.StatusMsg)

	}

	Addressid := types.AddressData{
		AddressID:       uint64(addaddressresp.Data.AddressId),
		RecipientName:   addaddressresp.Data.RecipientName,
		PhoneNumber:     addaddressresp.Data.PhoneNumber,
		Province:        addaddressresp.Data.Province,
		City:            addaddressresp.Data.City,
		DetailedAddress: addaddressresp.Data.DetailedAddress,
		IsDefault:       addaddressresp.Data.IsDefault,
		CreatedAt:       addaddressresp.Data.CreatedAt,
		UpdatedAt:       addaddressresp.Data.UpdatedAt,
	}

	resp = &types.AddAddressResponse{
		Data: Addressid,
	}

	return
}
