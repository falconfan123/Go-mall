package persistence

import (
	"context"
	"database/sql"

	"github.com/falconfan123/Go-mall/dal/model/audit"
	"github.com/falconfan123/Go-mall/services/audit/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/audit/internal/domain/repository"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// AuditRepositoryImpl 审计日志仓储实现
type AuditRepositoryImpl struct {
	conn       sqlx.SqlConn
	auditModel audit.AuditModel
}

// NewAuditRepositoryImpl 创建审计日志仓储实现
func NewAuditRepositoryImpl(conn sqlx.SqlConn) repository.AuditRepository {
	return &AuditRepositoryImpl{
		conn:       conn,
		auditModel: audit.NewAuditModel(conn),
	}
}

// GetByID 根据ID查询
func (r *AuditRepositoryImpl) GetByID(ctx context.Context, id uint64) (*entity.AuditLog, error) {
	a, err := r.auditModel.FindOne(ctx, id)
	if err != nil {
		if err == audit.ErrNotFound {
			return nil, entity.ErrAuditNotFound
		}
		return nil, err
	}

	return r.toDomain(a), nil
}

// GetByTraceID 根据TraceID查询
func (r *AuditRepositoryImpl) GetByTraceID(ctx context.Context, traceID string) (*entity.AuditLog, error) {
	a, err := r.auditModel.FindOneByTraceId(ctx, traceID)
	if err != nil {
		if err == audit.ErrNotFound {
			return nil, entity.ErrAuditNotFound
		}
		return nil, err
	}

	return r.toDomain(a), nil
}

// Save 保存审计日志
func (r *AuditRepositoryImpl) Save(ctx context.Context, auditLog *entity.AuditLog) error {
	auditDO := r.toDO(auditLog)
	_, err := r.auditModel.Insert(ctx, auditDO)
	return err
}

// Delete 删除审计日志
func (r *AuditRepositoryImpl) Delete(ctx context.Context, id uint64) error {
	return r.auditModel.Delete(ctx, id)
}

// ListByUserID 查询用户的审计日志列表
func (r *AuditRepositoryImpl) ListByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.AuditLog, int64, error) {
	// 简化实现：直接返回空列表
	// 实际应该使用分页查询
	return nil, 0, nil
}

// ListByTarget 查询目标对象的审计日志
func (r *AuditRepositoryImpl) ListByTarget(ctx context.Context, targetTable string, targetID uint64, page, pageSize int) ([]*entity.AuditLog, int64, error) {
	// 简化实现
	return nil, 0, nil
}

// ListByActionType 按操作类型查询
func (r *AuditRepositoryImpl) ListByActionType(ctx context.Context, actionType string, page, pageSize int) ([]*entity.AuditLog, int64, error) {
	// 简化实现
	return nil, 0, nil
}

// toDomain 转换数据库模型到领域模型
func (r *AuditRepositoryImpl) toDomain(a *audit.Audit) *entity.AuditLog {
	var actionDesc string
	if a.ActionDesc.Valid {
		actionDesc = a.ActionDesc.String
	}

	var oldData string
	if a.OldData.Valid {
		oldData = a.OldData.String
	}

	var newData string
	if a.NewData.Valid {
		newData = a.NewData.String
	}

	return &entity.AuditLog{
		ID:          a.Id,
		UserID:      a.UserId,
		ActionType:  a.ActionType,
		ActionDesc:  actionDesc,
		OldData:     oldData,
		NewData:     newData,
		ServiceName: a.ServiceName,
		TargetTable: a.TargetTable,
		TargetID:    a.TargetId,
		ClientIP:    a.ClientIp,
		TraceID:     a.TraceId,
		SpanID:      a.SpanId,
		CreatedAt:   a.CreatedAt,
	}
}

// toDO 转换领域模型到数据库模型
func (r *AuditRepositoryImpl) toDO(a *entity.AuditLog) *audit.Audit {
	var actionDesc, oldData, newData sql.NullString
	if a.ActionDesc != "" {
		actionDesc = sql.NullString{String: a.ActionDesc, Valid: true}
	}
	if a.OldData != "" {
		oldData = sql.NullString{String: a.OldData, Valid: true}
	}
	if a.NewData != "" {
		newData = sql.NullString{String: a.NewData, Valid: true}
	}

	return &audit.Audit{
		Id:          a.ID,
		UserId:      a.UserID,
		ActionType:  a.ActionType,
		ActionDesc:  actionDesc,
		OldData:     oldData,
		NewData:     newData,
		ServiceName: a.ServiceName,
		TargetTable: a.TargetTable,
		TargetId:    a.TargetID,
		ClientIp:    a.ClientIP,
		TraceId:     a.TraceID,
		SpanId:      a.SpanID,
		CreatedAt:   a.CreatedAt,
	}
}
