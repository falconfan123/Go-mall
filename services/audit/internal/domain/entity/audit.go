package entity

import (
	"errors"
	"time"
)

var (
	ErrAuditNotFound = errors.New("audit log not found")
)

// AuditLog 审计日志实体
type AuditLog struct {
	ID          uint64    // 主键
	UserID      uint64    // 用户ID
	ActionType  string    // 操作类型
	ActionDesc  string    // 操作描述
	OldData     string    // 旧数据
	NewData     string    // 新数据
	ServiceName string    // 服务名称
	TargetTable string    // 目标表
	TargetID    uint64    // 目标ID
	ClientIP    string    // 客户端IP
	TraceID     string    // Trace ID
	SpanID      string    // Span ID
	CreatedAt   time.Time // 创建时间
}

// NewAuditLog 创建审计日志
func NewAuditLog(
	userID uint64,
	actionType string,
	actionDesc string,
	oldData string,
	newData string,
	serviceName string,
	targetTable string,
	targetID uint64,
	clientIP string,
	traceID string,
	spanID string,
) *AuditLog {
	return &AuditLog{
		UserID:      userID,
		ActionType:  actionType,
		ActionDesc:  actionDesc,
		OldData:     oldData,
		NewData:     newData,
		ServiceName: serviceName,
		TargetTable: targetTable,
		TargetID:    targetID,
		ClientIP:    clientIP,
		TraceID:     traceID,
		SpanID:      spanID,
		CreatedAt:   time.Now(),
	}
}
