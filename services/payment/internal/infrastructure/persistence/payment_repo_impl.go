package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/falconfan123/Go-mall/dal/model/payment"
	"github.com/falconfan123/Go-mall/services/payment/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/payment/internal/domain/repository"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

// PaymentRepositoryImpl 支付单仓储实现
type PaymentRepositoryImpl struct {
	conn          sqlx.SqlConn
	paymentsModel payment.PaymentsModel
}

// NewPaymentRepositoryImpl 创建支付单仓储实现
func NewPaymentRepositoryImpl(conn sqlx.SqlConn) repository.PaymentRepository {
	return &PaymentRepositoryImpl{
		conn:          conn,
		paymentsModel: payment.NewPaymentsModel(conn),
	}
}

// GetByID 根据支付单ID查询
func (r *PaymentRepositoryImpl) GetByID(ctx context.Context, paymentID string) (*entity.Payment, error) {
	p, err := r.paymentsModel.FindOne(ctx, paymentID)
	if err != nil {
		if err == pb.ErrNotFound {
			return nil, entity.ErrPaymentNotFound
		}
		return nil, err
	}

	return r.toDomainPayment(p), nil
}

// GetByOrderID 根据订单ID查询
func (r *PaymentRepositoryImpl) GetByOrderID(ctx context.Context, orderID string) (*entity.Payment, error) {
	p, err := r.paymentsModel.FindOneByOrderId(ctx, orderID)
	if err != nil {
		if err == pb.ErrNotFound {
			return nil, entity.ErrPaymentNotFound
		}
		return nil, err
	}

	return r.toDomainPayment(p), nil
}

// GetByPreOrderID 根据预订单ID查询
func (r *PaymentRepositoryImpl) GetByPreOrderID(ctx context.Context, preOrderID string) (*entity.Payment, error) {
	// 先通过 pre_order_id 查询
	payments, err := r.paymentsModel.FindPage(ctx, 0, 0, 1000)
	if err != nil {
		return nil, err
	}

	for _, p := range payments {
		if p.PreOrderId == preOrderID {
			return r.toDomainPayment(p), nil
		}
	}

	return nil, entity.ErrPaymentNotFound
}

// Save 保存支付单
func (r *PaymentRepositoryImpl) Save(ctx context.Context, payment *entity.Payment) error {
	paymentDO := r.toDOPayment(payment)
	_, err := r.paymentsModel.Insert(ctx, paymentDO)
	return err
}

// Update 更新支付单
func (r *PaymentRepositoryImpl) Update(ctx context.Context, payment *entity.Payment) error {
	paymentDO := r.toDOPayment(payment)
	return r.paymentsModel.Update(ctx, paymentDO)
}

// Delete 删除支付单
func (r *PaymentRepositoryImpl) Delete(ctx context.Context, paymentID string) error {
	return r.paymentsModel.Delete(ctx, paymentID)
}

// ListByUserID 查询用户的支付单列表
func (r *PaymentRepositoryImpl) ListByUserID(ctx context.Context, userID int64, status *entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int64, error) {
	offset := (page - 1) * pageSize
	payments, err := r.paymentsModel.FindPage(ctx, uint32(userID), offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 如果有状态过滤
	result := make([]*entity.Payment, 0, len(payments))
	for _, p := range payments {
		domainPayment := r.toDomainPayment(p)
		if status != nil && domainPayment.Status != *status {
			continue
		}
		result = append(result, domainPayment)
	}

	// 统计总数
	total, err := r.paymentsModel.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// ListByStatus 根据状态查询支付单列表
func (r *PaymentRepositoryImpl) ListByStatus(ctx context.Context, status entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int64, error) {
	// 简化实现：查询所有然后过滤
	payments, err := r.paymentsModel.FindPage(ctx, 0, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*entity.Payment, 0, len(payments))
	for _, p := range payments {
		domainPayment := r.toDomainPayment(p)
		if domainPayment.Status == status {
			result = append(result, domainPayment)
		}
	}

	return result, int64(len(result)), nil
}

// FindExpired 查找已过期的支付单
func (r *PaymentRepositoryImpl) FindExpired(ctx context.Context, limit int) ([]*entity.Payment, error) {
	payments, err := r.paymentsModel.FindExpired(ctx, limit)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Payment, 0, len(payments))
	for _, p := range payments {
		result = append(result, r.toDomainPayment(p))
	}

	return result, nil
}

// toDomainPayment 转换数据库模型到领域模型
func (r *PaymentRepositoryImpl) toDomainPayment(p *payment.Payments) *entity.Payment {
	var paidAt *time.Time
	if p.PaidAt.Valid {
		t := time.Unix(p.PaidAt.Int64, 0)
		paidAt = &t
	}

	var paidAmount int64
	if p.PaidAmount.Valid {
		paidAmount = p.PaidAmount.Int64
	}

	return &entity.Payment{
		PaymentID:      p.PaymentId,
		PreOrderID:     p.PreOrderId,
		OrderID:        p.OrderId.String,
		UserID:         int64(p.UserId),
		OriginalAmount: p.OriginalAmount,
		PaidAmount:     paidAmount,
		PaymentMethod:  entity.PaymentMethodFromString(p.PaymentMethod),
		TransactionID:  p.TransactionId.String,
		PayURL:         p.PayUrl,
		ExpireTime:     time.Unix(p.ExpireTime, 0),
		Status:         entity.PaymentStatus(p.Status),
		PaidAt:         paidAt,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// toDOPayment 转换领域模型到数据库模型
func (r *PaymentRepositoryImpl) toDOPayment(p *entity.Payment) *payment.Payments {
	var orderID sql.NullString
	if p.OrderID != "" {
		orderID = sql.NullString{String: p.OrderID, Valid: true}
	}

	var paidAmount sql.NullInt64
	if p.PaidAmount > 0 {
		paidAmount = sql.NullInt64{Int64: p.PaidAmount, Valid: true}
	}

	var transactionID sql.NullString
	if p.TransactionID != "" {
		transactionID = sql.NullString{String: p.TransactionID, Valid: true}
	}

	var paidAt sql.NullInt64
	if p.PaidAt != nil {
		paidAt = sql.NullInt64{Int64: p.PaidAt.Unix(), Valid: true}
	}

	return &payment.Payments{
		PaymentId:      p.PaymentID,
		PreOrderId:     p.PreOrderID,
		OrderId:        orderID,
		UserId:         uint64(p.UserID),
		OriginalAmount: p.OriginalAmount,
		PaidAmount:     paidAmount,
		PaymentMethod:  p.PaymentMethod.String(),
		TransactionId:  transactionID,
		PayUrl:         p.PayURL,
		ExpireTime:     p.ExpireTime.Unix(),
		Status:         int64(p.Status),
		PaidAt:         paidAt,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
