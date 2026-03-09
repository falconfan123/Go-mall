package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"jijizhazha1024/go-mall/common/consts/code"
	"jijizhazha1024/go-mall/services/carts/carts"
	"jijizhazha1024/go-mall/services/carts/internal/svc"
	"strconv"

	"google.golang.org/grpc/metadata"
)

type CartItemListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCartItemListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CartItemListLogic {
	return &CartItemListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CartItemListLogic) CartItemList(in *carts.UserInfo) (*carts.CartItemListResponse, error) {
	// Read user_id from metadata
	md, ok := metadata.FromIncomingContext(l.ctx)
	if ok {
		uis := md.Get("user_id")
		if len(uis) > 0 {
			uid, _ := strconv.Atoi(uis[0])
			if uid > 0 {
				in.Id = int32(uid)
			}
		}
	}

	// 定义响应对象
	var rsp carts.CartItemListResponse

	shopCarts, err := l.svcCtx.CartsModel.FindByUserID(l.ctx, int64(in.Id))
	if err != nil {
		l.Logger.Errorw("get shopCarts from database failed",
			logx.Field("err", err),
			logx.Field("user_id", in.Id))
		return &carts.CartItemListResponse{
			StatusCode: code.CartInfoRetrievalFailed,
			StatusMsg:  code.CartInfoRetrievalFailedMsg,
			Total:      0,
			Data:       nil,
		}, err
	}

	// 设置响应中的总数
	rsp.Total = int32(len(shopCarts))

	// 构建响应数据
	for _, shopCart := range shopCarts {
		rsp.Data = append(rsp.Data, &carts.CartInfoResponse{
			Id:        int32(shopCart.Id),
			UserId:    int32(shopCart.UserId.Int64),
			ProductId: int32(shopCart.ProductId.Int64),
			Quantity:  int32(shopCart.Quantity.Int64),
			Checked:   shopCart.Checked.Int64 == 1,
		})
	}

	return &carts.CartItemListResponse{
		StatusCode: code.Success,
		StatusMsg:  code.CartInfoRetrievedMsg,
		Total:      rsp.Total,
		Data:       rsp.Data,
	}, nil
}
