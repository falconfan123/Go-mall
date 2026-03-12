package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/apis/product/internal/svc"
	"github.com/falconfan123/Go-mall/apis/product/internal/types"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

// GetProductByIDLogic is the business logic for GetProductByIDLogic operations.
type GetProductByIDLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetProductByIDLogic creates a new GetProductByIDLogic instance.
func NewGetProductByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductByIDLogic {
	return &GetProductByIDLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// does something.
func (l *GetProductByIDLogic) GetProductByID(req *types.GetProductByIDReq) (resp *types.GetProductByIDResp, err error) {
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, errors.New(code.AuthBlank, code.AuthBlankMsg)
	}
	res, err := l.svcCtx.ProductRPC.GetProduct(l.ctx, &product.GetProductReq{
		Id:     uint32(req.ID),
		UserId: int32(userID),
	})
	if err != nil {
		l.Logger.Errorf("call rpc ProductRPC.GetProduct failed", logx.Field("err", err))
		return nil, errors.New(int(res.StatusCode), res.StatusMsg)
	}
	if res.StatusCode != code.Success {
		// 提示用户
		return nil, errors.New(int(res.StatusCode), res.StatusMsg)
	}

	resp = &types.GetProductByIDResp{
		ID:          int64(res.Product.Id),
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Picture:     res.Product.Picture,
		Stock:       res.Product.Stock,
		Price:       res.Product.Price,
		Sold:        res.Product.Sold,
		Categories:  res.Product.Categories,
		CreatedAt:   res.Product.CratedAt,
		UpdatedAt:   res.Product.UpdatedAt,
	}

	return resp, err
}
