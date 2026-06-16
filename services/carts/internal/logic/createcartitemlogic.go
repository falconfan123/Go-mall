package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/carts/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/carts/internal/svc"
	carts "github.com/falconfan123/Go-mall/services/carts/pb"

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
	md, ok := metadata.FromIncomingContext(l.ctx)
	if ok {
		if userID := userIDFromMetadata(md); userID > 0 {
			in.UserId = userID
		}
	}

	// 1. 转换为DTO
	req := &dto.AddItemReq{
		UserID:       int64(in.UserId),
		ProductID:    int64(in.ProductId),
		ProductName:  in.ProductName,
		ProductImage: in.ProductImage,
		ProductPrice: float64(in.ProductPrice),
		Quantity:     in.Quantity + 1, // 兼容原有逻辑，数量+1
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
