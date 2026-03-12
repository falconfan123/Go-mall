package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/code"
	carts "github.com/falconfan123/Go-mall/services/carts/pb"
	"github.com/falconfan123/Go-mall/services/carts/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/carts/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
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

	// 1. 转换为DTO
	req := &dto.GetCartReq{
		UserID: int64(in.Id),
	}

	// 2. 调用应用服务
	cartDTO, err := l.svcCtx.CartAppService.GetCart(l.ctx, req)
	if err != nil {
		l.Logger.Errorw("Failed to get cart",
			logx.Field("err", err),
			logx.Field("user_id", in.Id))
		return &carts.CartItemListResponse{
			StatusCode: code.CartInfoRetrievalFailed,
			StatusMsg:  code.CartInfoRetrievalFailedMsg,
			Total:      0,
			Data:       nil,
		}, err
	}

	// 3. 转换为Proto响应
	var rsp carts.CartItemListResponse
	rsp.Total = int32(len(cartDTO.Items))

	for _, item := range cartDTO.Items {
		rsp.Data = append(rsp.Data, &carts.CartInfoResponse{
			ProductId: int32(item.ProductID),
			Quantity:  item.Quantity,
			Checked:   item.Checked,
			// 注意：原有响应缺少商品名称、图片、价格字段，可根据需要扩展
		})
	}

	return &carts.CartItemListResponse{
		StatusCode: code.Success,
		StatusMsg:  code.CartInfoRetrievedMsg,
		Total:      rsp.Total,
		Data:       rsp.Data,
	}, nil
}
