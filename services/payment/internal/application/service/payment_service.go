package service

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/payment/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/payment/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/payment/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/payment/internal/domain/repository"
)

// PaymentAppService 支付应用服务
type PaymentAppService struct {
	paymentRepo repository.PaymentRepository
}

// NewPaymentAppService 创建支付应用服务
func NewPaymentAppService(paymentRepo repository.PaymentRepository) *PaymentAppService {
	return &PaymentAppService{
		paymentRepo: paymentRepo,
	}
}

// CreatePayment 创建支付单
func (s *PaymentAppService) CreatePayment(ctx context.Context, req *dto.CreatePaymentReq) (*dto.CreatePaymentResp, error) {
	// 1. 幂等性校验：根据订单ID查询是否已经创建过支付单
	existingPayment, err := s.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil && err != entity.ErrPaymentNotFound {
		return &dto.CreatePaymentResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	if existingPayment != nil {
		return &dto.CreatePaymentResp{
			Payment:    s.convertPaymentToDTO(existingPayment),
			StatusCode: code.PaymentExist,
			StatusMsg:  code.PaymentExistMsg,
		}, nil
	}

	// 2. 创建支付单（实际逻辑需要在外部调用订单服务获取订单信息）
	// 这里简化处理，调用方需要传入必要信息
	paymentID := s.generatePaymentID()

	var paymentMethod entity.PaymentMethod
	switch req.PaymentMethod {
	case 1:
		paymentMethod = entity.PaymentMethodAlipay
	case 2:
		paymentMethod = entity.PaymentMethodWechat
	default:
		return &dto.CreatePaymentResp{
			StatusCode: code.PaymentMethodNotSupport,
			StatusMsg:  code.PaymentMethodNotSupportMsg,
		}, nil
	}

	// 创建聚合根
	paymentAgg := aggregate.NewPaymentAggregate(
		paymentID,
		"",          // preOrderID - 由调用方填充
		req.OrderID, // orderID
		req.UserID,  // userID
		0,           // originalAmount - 由调用方填充
		0,           // paidAmount - 由调用方填充
		paymentMethod,
		"", // payURL - 由调用方生成
		30, // 30分钟过期
	)

	// 保存
	payment := paymentAgg.GetPayment()
	if err := s.paymentRepo.Save(ctx, payment); err != nil {
		return &dto.CreatePaymentResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.CreatePaymentResp{
		Payment:    s.convertPaymentToDTO(payment),
		StatusCode: code.Success,
		StatusMsg:  "payment created successfully",
	}, nil
}

// GetPayment 获取支付单详情
func (s *PaymentAppService) GetPayment(ctx context.Context, req *dto.GetPaymentReq) (*dto.GetPaymentResp, error) {
	payment, err := s.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		if err == entity.ErrPaymentNotFound {
			return &dto.GetPaymentResp{
				StatusCode: code.PaymentNotExist,
				StatusMsg:  code.PaymentNotExistMsg,
			}, nil
		}
		return &dto.GetPaymentResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.GetPaymentResp{
		Payment:    s.convertPaymentToDTO(payment),
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// ListPayments 查询支付列表
func (s *PaymentAppService) ListPayments(ctx context.Context, req *dto.ListPaymentsReq) (*dto.ListPaymentsResp, error) {
	var status *entity.PaymentStatus
	if req.Status != nil {
		st := entity.PaymentStatus(*req.Status)
		status = &st
	}

	payments, total, err := s.paymentRepo.ListByUserID(ctx, req.UserID, status, req.Page, req.PageSize)
	if err != nil {
		return &dto.ListPaymentsResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	items := make([]*dto.PaymentListItemDTO, 0, len(payments))
	for _, p := range payments {
		items = append(items, &dto.PaymentListItemDTO{
			PaymentID:     p.PaymentID,
			OrderID:       p.OrderID,
			PaidAmount:    p.PaidAmount,
			Status:        int(p.Status),
			PaymentMethod: s.paymentMethodToInt(p.PaymentMethod),
			CreatedAt:     p.CreatedAt.Unix(),
		})
	}

	return &dto.ListPaymentsResp{
		Payments:   items,
		TotalCount: total,
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// 辅助方法
func (s *PaymentAppService) generatePaymentID() string {
	return fmt.Sprintf("PAY%d%d", time.Now().Unix(), time.Now().Nanosecond())
}

func (s *PaymentAppService) convertPaymentToDTO(payment *entity.Payment) *dto.PaymentDTO {
	var paidAt *int64
	if payment.PaidAt != nil {
		pt := payment.PaidAt.Unix()
		paidAt = &pt
	}

	return &dto.PaymentDTO{
		PaymentID:      payment.PaymentID,
		PreOrderID:     payment.PreOrderID,
		OrderID:        payment.OrderID,
		UserID:         payment.UserID,
		OriginalAmount: payment.OriginalAmount,
		PaidAmount:     payment.PaidAmount,
		PaymentMethod:  s.paymentMethodToInt(payment.PaymentMethod),
		TransactionID:  payment.TransactionID,
		PayURL:         payment.PayURL,
		Status:         int(payment.Status),
		ExpireTime:     payment.ExpireTime.Unix(),
		PaidAt:         paidAt,
		CreatedAt:      payment.CreatedAt.Unix(),
		UpdatedAt:      payment.UpdatedAt.Unix(),
	}
}

func (s *PaymentAppService) paymentMethodToInt(method entity.PaymentMethod) int {
	switch method {
	case entity.PaymentMethodAlipay:
		return 1
	case entity.PaymentMethodWechat:
		return 2
	default:
		return 0
	}
}

func init() {
	_ = code.ServerError // 引入code包
}
