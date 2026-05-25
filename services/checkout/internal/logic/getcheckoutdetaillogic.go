package logic

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/checkout/internal/svc"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCheckoutDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCheckoutDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCheckoutDetailLogic {
	return &GetCheckoutDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetCheckoutDetail 获取结算详情
func (l *GetCheckoutDetailLogic) GetCheckoutDetail(in *checkout.CheckoutDetailReq) (*checkout.CheckoutDetailResp, error) {
	checkoutRecord, err := l.svcCtx.CheckoutModel.FindOneByUserIdAndPreOrderId(l.ctx, in.UserId, in.PreOrderId)
	res := &checkout.CheckoutDetailResp{}
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			res.StatusCode = code.OutOfRecord
			res.StatusMsg = code.OutOfRecordMsg
			return res, nil
		} else {
			l.Logger.Errorw("查询结算记录失败",
				logx.Field("err", err),
				logx.Field("user_id", in.UserId),
				logx.Field("pre_order_id", in.PreOrderId))
			return nil, err
		}
	}

	checkoutItems, err := l.svcCtx.CheckoutItemsModel.FindItemsByPreOrder(l.ctx, in.PreOrderId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			res.StatusCode = code.OutOfRecord
			res.StatusMsg = code.OutOfRecordMsg
			return res, nil
		} else {
			l.Logger.Errorw("查询结算记录失败",
				logx.Field("err", err),
				logx.Field("user_id", in.UserId),
				logx.Field("pre_order_id", in.PreOrderId))
			return nil, err
		}
	}

	orderData := &checkout.CheckoutOrder{
		PreOrderId:     checkoutRecord.PreOrderId,
		UserId:         int64(checkoutRecord.UserId),
		Status:         checkout.CheckoutStatus(checkoutRecord.Status),
		ExpireTime:     checkoutRecord.ExpireTime,
		CreatedAt:      checkoutRecord.CreatedAt.Format(time.DateTime),
		UpdatedAt:      checkoutRecord.UpdatedAt.Format(time.DateTime),
		FinalAmount:    checkoutRecord.FinalAmount,
		OriginalAmount: checkoutRecord.OriginalAmount,
	}

	var items []*checkout.CheckoutItem
	for _, item := range checkoutItems {
		var snapshot map[string]string
		if err := json.Unmarshal([]byte(item.Snapshot), &snapshot); err != nil {
			l.Logger.Errorw("解析 checkout item snapshot 失败",
				logx.Field("err", err),
				logx.Field("pre_order_id", in.PreOrderId),
				logx.Field("product_id", item.ProductId))
			snapshot = map[string]string{}
		}
		checkoutItem := &checkout.CheckoutItem{
			ProductId:   int32(item.ProductId),
			Quantity:    int32(item.Quantity),
			Price:       item.Price,
			ProductName: snapshot["name"],
			ProductDesc: snapshot["desc"],
		}
		items = append(items, checkoutItem)
	}

	orderData.Items = items

	resp := &checkout.CheckoutDetailResp{
		StatusCode: code.Success,
		StatusMsg:  code.SuccessMsg,
		Data: orderData,
	}

	return resp, nil
}
