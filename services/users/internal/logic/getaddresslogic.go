package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/dal/model/user_address"
	"github.com/falconfan123/Go-mall/services/users/internal/svc"
	"github.com/falconfan123/Go-mall/services/users/userspb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个收货地址
func (l *GetAddressLogic) GetAddress(in *userspb.GetAddressRequest) (*userspb.GetAddressResponse, error) {
	address, err := l.svcCtx.UserAddressesModel.FindOne(l.ctx, in.AddressId)
	if err != nil {
		if err == user_address.ErrNotFound {
			return &userspb.GetAddressResponse{
				StatusCode: int32(code.AddressNotExist),
				StatusMsg:  code.AddressNotExistMsg,
			}, nil
		}
		l.Logger.Errorw("get address failed", logx.Field("err", err))
		return &userspb.GetAddressResponse{
			StatusCode: int32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, nil
	}

	if uint32(address.UserId) != in.UserId {
		return &userspb.GetAddressResponse{
			StatusCode: int32(code.AddressNotExist), // Or PermissionDenied
			StatusMsg:  code.AddressNotExistMsg,
		}, nil
	}

	return &userspb.GetAddressResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		Data: &userspb.AddressData{
			AddressId:       uint64(address.AddressId),
			RecipientName:   address.RecipientName,
			PhoneNumber:     address.PhoneNumber.String,
			Province:        address.Province.String,
			City:            address.City,
			DetailedAddress: address.DetailedAddress,
			IsDefault:       address.IsDefault,
			CreatedAt:       address.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       address.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
