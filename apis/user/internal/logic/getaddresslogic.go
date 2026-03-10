package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/apis/user/internal/svc"
	"github.com/falconfan123/Go-mall/apis/user/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/users/users"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type GetAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAddressLogic) GetAddress(req *types.GetAddressRequest) (resp *types.GetAddressResponse, err error) {

	user_id := l.ctx.Value(biz.UserIDKey).(uint32)
	getaddressresp, err := l.svcCtx.UserRpc.GetAddress(l.ctx, &users.GetAddressRequest{
		UserId:    user_id,
		AddressId: req.AddressID,
	})

	if err != nil {
		l.Logger.Errorw("调用 rpc 获取地址失败", logx.Field("err", err))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	} else if getaddressresp.StatusMsg != "" {

		return nil, errors.New(int(getaddressresp.StatusCode), getaddressresp.StatusMsg)

	}

	// 创建响应对象并填充数据
	resp = &types.GetAddressResponse{
		Data: types.AddressData{
			AddressID:       uint64(getaddressresp.Data.AddressId),
			RecipientName:   getaddressresp.Data.RecipientName,
			PhoneNumber:     getaddressresp.Data.PhoneNumber,
			Province:        getaddressresp.Data.Province,
			City:            getaddressresp.Data.City,
			DetailedAddress: getaddressresp.Data.DetailedAddress,
			IsDefault:       getaddressresp.Data.IsDefault,
			CreatedAt:       getaddressresp.Data.CreatedAt,
			UpdatedAt:       getaddressresp.Data.UpdatedAt,
		},
	}

	return resp, nil

}
