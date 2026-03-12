package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/carts/pb"
	"github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/zeromicro/x/errors"

	"github.com/falconfan123/Go-mall/apis/carts/internal/svc"
	"github.com/falconfan123/Go-mall/apis/carts/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreateCartItemLogic is the business logic for CreateCartItemLogic operations.
type CreateCartItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateCartItemLogic creates a new CreateCartItemLogic instance.
func NewCreateCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCartItemLogic {
	return &CreateCartItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// does something.
func (l *CreateCartItemLogic) CreateCartItem(req *types.CreateCartReq) (resp *types.CreateCartResp, err error) {
	userID := l.ctx.Value(biz.UserIDKey).(uint32)

	// 1. 先检查商品是否存在
	exist, err := l.svcCtx.ProductRPC.IsExistProduct(l.ctx, &product.IsExistProductReq{
		Id: int64(req.ProductId),
	})
	if err != nil {
		l.Logger.Errorw("call rpc IsExistProduct failed",
			logx.Field("err", err),
			logx.Field("product_id", req.ProductId))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	}
	if !exist.Exist {
		l.Logger.Errorw("product does not exist",
			logx.Field("product_id", req.ProductId))
		return nil, errors.New(code.ProductInfoRetrievalFailed, code.ProductInfoRetrievalFailedMsg)
	}

	// 2. 调用 GetProduct RPC 获取商品详情
	productRes, err := l.svcCtx.ProductRPC.GetProduct(l.ctx, &product.GetProductReq{
		Id: uint32(req.ProductId),
	})
	if err != nil {
		l.Logger.Errorw("call rpc GetProduct failed",
			logx.Field("err", err),
			logx.Field("product_id", req.ProductId))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	}

	// 3. 检查库存是否足够
	if productRes.Product.Stock == 0 {
		l.Logger.Errorw("insufficient stock",
			logx.Field("product_id", req.ProductId),
			logx.Field("available_stock", productRes.Product.Stock))
		return nil, errors.New(code.InsufficientInventoryOfProduct, code.InsufficientInventoryOfProductMsg)
	}

	// 4. 调用 CreateCartItem RPC 添加到购物车
	res, err := l.svcCtx.CartsRPC.CreateCartItem(l.ctx, &carts.CartItemRequest{
		UserId:    int32(userID),
		ProductId: req.ProductId,
	})
	if err != nil {
		l.Logger.Errorw("call rpc CreateCartItem failed",
			logx.Field("err", err),
			logx.Field("userID", userID),
			logx.Field("product_id", req.ProductId))
		return nil, errors.New(code.CartCreationFailed, code.CartCreationFailedMsg)
	}

	// 5. 处理 RPC 返回 nil 的情况
	if res == nil {
		l.Logger.Errorw("rpc CreateCartItem returned nil response",
			logx.Field("userID", userID),
			logx.Field("product_id", req.ProductId))
		return nil, errors.New(code.ServerError, code.ServerErrorMsg)
	}

	// 6. 处理业务错误
	if res.StatusCode != code.Success {
		l.Logger.Debugw("rpc CreateCartItem returned business error",
			logx.Field("userID", userID),
			logx.Field("product_id", req.ProductId),
			logx.Field("status_code", res.StatusCode),
			logx.Field("status_msg", res.StatusMsg))
		return nil, errors.New(int(res.StatusCode), res.StatusMsg)
	}

	// 7. 记录成功日志并返回结果
	l.Logger.Infow("Cart item created successfully",
		logx.Field("userID", userID),
		logx.Field("product_id", req.ProductId))

	return &types.CreateCartResp{
		Id: res.Id,
	}, nil
}
