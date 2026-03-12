package dto

// CreateAuditLogReq 创建审计日志请求
type CreateAuditLogReq struct {
	UserID      uint64 `json:"userId"`
	ActionType  string `json:"actionType"`
	ActionDesc  string `json:"actionDesc,omitempty"`
	OldData     string `json:"oldData,omitempty"`
	NewData     string `json:"newData,omitempty"`
	ServiceName string `json:"serviceName"`
	TargetTable string `json:"targetTable"`
	TargetID    uint64 `json:"targetId"`
	ClientIP    string `json:"clientIp"`
	TraceID     string `json:"traceId,omitempty"`
	SpanID      string `json:"spanId,omitempty"`
}

// CreateAuditLogResp 创建审计日志响应
type CreateAuditLogResp struct {
	OK         bool   `json:"ok"`
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// GetAuditLogReq 获取审计日志请求
type GetAuditLogReq struct {
	ID uint64 `json:"id"`
}

// GetAuditLogResp 获取审计日志响应
type GetAuditLogResp struct {
	AuditLog   *AuditLogDTO `json:"auditLog"`
	StatusCode int64        `json:"statusCode"`
	StatusMsg  string       `json:"statusMsg"`
}

// ListAuditLogsReq 审计日志列表请求
type ListAuditLogsReq struct {
	UserID      uint64 `json:"userId,omitempty"`
	ActionType  string `json:"actionType,omitempty"`
	TargetTable string `json:"targetTable,omitempty"`
	TargetID    uint64 `json:"targetId,omitempty"`
	Page        int    `json:"page"`
	PageSize    int    `json:"pageSize"`
}

// ListAuditLogsResp 审计日志列表响应
type ListAuditLogsResp struct {
	AuditLogs  []*AuditLogDTO `json:"auditLogs"`
	TotalCount int64          `json:"totalCount"`
	StatusCode int64          `json:"statusCode"`
	StatusMsg  string         `json:"statusMsg"`
}

// AuditLogDTO 审计日志DTO
type AuditLogDTO struct {
	ID          uint64 `json:"id"`
	UserID      uint64 `json:"userId"`
	ActionType  string `json:"actionType"`
	ActionDesc  string `json:"actionDesc,omitempty"`
	OldData     string `json:"oldData,omitempty"`
	NewData     string `json:"newData,omitempty"`
	ServiceName string `json:"serviceName"`
	TargetTable string `json:"targetTable"`
	TargetID    uint64 `json:"targetId"`
	ClientIP    string `json:"clientIp"`
	TraceID     string `json:"traceId,omitempty"`
	SpanID      string `json:"spanId,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
}
