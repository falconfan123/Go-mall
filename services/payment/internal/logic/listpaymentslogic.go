package logic

import (
	"context"
	"errors"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	payment "github.com/falconfan123/Go-mall/services/payment/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPaymentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPaymentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPaymentsLogic {
	return &ListPaymentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPaymentsLogic) ListPayments(in *payment.PaymentListReq) (*payment.PaymentListResp, error) {
	var queryErr error
	paymentModel := l.svcCtx.PaymentModel
	page := int32(1)
	pageSize := int32(100)
	if in.Pagination != nil {
		if in.Pagination.Page > 0 {
			page = in.Pagination.Page
		}
		if in.Pagination.PageSize > 0 {
			pageSize = in.Pagination.PageSize
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	payments, queryErr := paymentModel.FindPage(l.ctx, in.UserId, int(offset), int(pageSize))
	// 统一错误处理
	if queryErr != nil {
		if errors.Is(queryErr, sqlx.ErrNotFound) {
			return &payment.PaymentListResp{}, nil
		}
		l.Logger.Errorw("query payments failed",
			logx.Field("err", queryErr),
			logx.Field("page", in.Pagination.Page),
			logx.Field("pageSize", in.Pagination.PageSize))
		return nil, queryErr
	}
	items := make([]*payment.PaymentItem, len(payments))
	for i, p := range payments {
		items[i] = &payment.PaymentItem{
			PaymentId:      p.PaymentId,
			PreOrderId:     p.PreOrderId,
			OrderId:        p.OrderId.String,
			OriginalAmount: p.OriginalAmount,
			PaidAmount:     p.PaidAmount.Int64,
			PaymentMethod:  paymentMethodEnum(p.PaymentMethod),
			TransactionId:  p.TransactionId.String,
			PayUrl:         p.PayUrl.String,
			ExpireTime:     p.ExpireTime,
			Status:         payment.PaymentStatus(p.Status),
			CreatedAt:      p.CreatedAt.Unix(),
			UpdatedAt:      p.UpdatedAt.Unix(),
			PaidAt:         p.PaidAt.Int64,
		}
	}

	return &payment.PaymentListResp{
		Payments: items,
	}, nil
}

func paymentMethodEnum(method string) payment.PaymentMethod {
	if method == "stripe" {
		return payment.PaymentMethod_STRIPE
	}
	return payment.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
}
