package logic

import (
	"context"
	"database/sql"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/dal/model/user_address"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/userspb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAddressLogic {
	return &AddAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddAddressLogic) AddAddress(in *userspb.AddAddressRequest) (*userspb.AddAddressResponse, error) {
	newAddress := &user_address.UserAddresses{
		UserId:          int64(in.UserId),
		RecipientName:   in.RecipientName,
		PhoneNumber:     sql.NullString{String: in.PhoneNumber, Valid: in.PhoneNumber != ""},
		Province:        sql.NullString{String: in.Province, Valid: in.Province != ""},
		City:            in.City,
		DetailedAddress: in.DetailedAddress,
		IsDefault:       in.IsDefault, // bool type
	}

	res, err := l.svcCtx.UserAddressesModel.Insert(l.ctx, newAddress)
	if err != nil {
		l.Logger.Errorw("add address failed", logx.Field("err", err))
		return &userspb.AddAddressResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	id, _ := res.LastInsertId()

	return &userspb.AddAddressResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		Data: &userspb.AddressData{
			AddressId:       uint64(id),
			RecipientName:   in.RecipientName,
			PhoneNumber:     in.PhoneNumber,
			Province:        in.Province,
			City:            in.City,
			DetailedAddress: in.DetailedAddress,
			IsDefault:       in.IsDefault,
		},
	}, nil
}
