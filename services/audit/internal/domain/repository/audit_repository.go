package repository

import (
	"context"

	"github.com/falconfan123/Go-mall/services/audit/internal/domain/entity"
)

// AuditRepository 审计日志仓储接口
type AuditRepository interface {
	// GetByID 根据ID查询
	GetByID(ctx context.Context, id uint64) (*entity.AuditLog, error)

	// GetByTraceID 根据TraceID查询
	GetByTraceID(ctx context.Context, traceID string) (*entity.AuditLog, error)

	// Save 保存审计日志
	Save(ctx context.Context, auditLog *entity.AuditLog) error

	// Delete 删除审计日志
	Delete(ctx context.Context, id uint64) error

	// ListByUserID 查询用户的审计日志列表
	ListByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.AuditLog, int64, error)

	// ListByTarget 查询目标对象的审计日志
	ListByTarget(ctx context.Context, targetTable string, targetID uint64, page, pageSize int) ([]*entity.AuditLog, int64, error)

	// ListByActionType 按操作类型查询
	ListByActionType(ctx context.Context, actionType string, page, pageSize int) ([]*entity.AuditLog, int64, error)
}
