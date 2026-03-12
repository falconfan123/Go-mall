package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/carts/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/carts/internal/svc"
	carts "github.com/falconfan123/Go-mall/services/carts/pb"
	"strconv"

	"google.golang.org/grpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCartItemLogic {
	return &CreateCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCartItemLogic) CreateCartItem(in *carts.CartItemRequest) (*carts.CreateCartResponse, error) {
	// Read user_id from metadata
	md, ok := metadata.FromIncomingContext(l.ctx)
	if ok {
		uis := md.Get("user_id")
		if len(uis) > 0 {
			uid, _ := strconv.Atoi(uis[0])
			if uid > 0 {
				in.UserId = int32(uid)
			}
		}
	}

	// 1. 转换为DTO
	req := &dto.AddItemReq{
		UserID:    int64(in.UserId),
		ProductID: int64(in.ProductId),
		Quantity:  in.Quantity + 1, // 兼容原有逻辑，数量+1
		// 注意：商品名称、图片、价格需要从商品服务获取
		// 这里简化处理，实际应该调用商品服务获取完整信息
		ProductName:  "",
		ProductImage: "",
		ProductPrice: 0,
	}

	// 2. 调用应用服务
	err := l.svcCtx.CartAppService.AddItem(l.ctx, req)
	if err != nil {
		l.Logger.Errorw("Failed to add item to cart",
			logx.Field("err", err),
			logx.Field("user_id", in.UserId),
			logx.Field("product_id", in.ProductId))
		return &carts.CreateCartResponse{
			StatusCode: code.CartCreationFailed,
			StatusMsg:  code.CartCreationFailedMsg,
			Id:         0,
		}, err
	}

	// 3. 返回响应
	return &carts.CreateCartResponse{
		StatusCode: code.Success,
		StatusMsg:  code.CartCreatedMsg,
		Id:         0, // 实际可以返回商品ID
	}, nil
}
