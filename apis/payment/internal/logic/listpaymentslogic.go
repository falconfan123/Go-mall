package logic

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/payment/pb"
	xerrors "github.com/zeromicro/x/errors"

	"github.com/falconfan123/Go-mall/apis/payment/internal/svc"
	"github.com/falconfan123/Go-mall/apis/payment/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListPaymentsLogic is the business logic for listpayments operations.
// ListPaymentsLogic is the business logic for ListPaymentsLogic operations.
type ListPaymentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListPaymentsLogic creates a new instance.
// NewListPaymentsLogic creates a new ListPaymentsLogic instance.
func NewListPaymentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPaymentsLogic {
	return &ListPaymentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListPayments is a function.
//
//	does something.
func (l *ListPaymentsLogic) ListPayments(req *types.PaymentListReq) (resp *types.PaymentListResponse, err error) {
	userID, ok := l.ctx.Value(biz.UserIDKey).(uint32)
	if !ok {
		return nil, xerrors.New(code.AuthBlank, code.AuthBlankMsg)
	}

	res, err := l.svcCtx.PaymentRPC.ListPayments(l.ctx, &payment.PaymentListReq{
		UserId: userID,
		Pagination: &payment.PaymentListReq_Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		PaymentMethod: payment.PaymentMethod_ALIPAY,
	})
	if err != nil {
		l.Logger.Errorw("call rpc ListPayments failed", logx.Field("err", err))
		return nil, xerrors.New(code.ServerError, code.ServerErrorMsg)
	}
	if res.StatusCode != code.Success {
		return nil, xerrors.New(int(res.StatusCode), res.StatusMsg)
	}
	resp = &types.PaymentListResponse{}
	resp.Data = make([]types.PaymentItem, len(res.Payments))
	for i, item := range res.Payments {
		resp.Data[i] = types.PaymentItem{
			PaymentID:     item.PaymentId,
			OrderID:       item.OrderId,
			PaidAmount:    item.PaidAmount,
			PayURL:        item.PayUrl,
			PaymentMethod: int32(item.PaymentMethod),
			Status:        int32(item.Status),
			TransactionID: item.TransactionId,
			CreatedAt:     item.CreatedAt,
		}
	}
	return
}
