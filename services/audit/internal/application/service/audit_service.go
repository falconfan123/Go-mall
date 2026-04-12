package service

import (
	"context"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/audit/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/audit/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/audit/internal/domain/repository"
)

// AuditAppService 审计应用服务
type AuditAppService struct {
	auditRepo repository.AuditRepository
}

// NewAuditAppService 创建审计应用服务
func NewAuditAppService(auditRepo repository.AuditRepository) *AuditAppService {
	return &AuditAppService{
		auditRepo: auditRepo,
	}
}

// CreateAuditLog 创建审计日志
func (s *AuditAppService) CreateAuditLog(ctx context.Context, req *dto.CreateAuditLogReq) (*dto.CreateAuditLogResp, error) {
	// 创建审计日志实体
	auditLog := entity.NewAuditLog(
		req.UserID,
		req.ActionType,
		req.ActionDesc,
		req.OldData,
		req.NewData,
		req.ServiceName,
		req.TargetTable,
		req.TargetID,
		req.ClientIP,
		req.TraceID,
		req.SpanID,
	)

	// 保存
	if err := s.auditRepo.Save(ctx, auditLog); err != nil {
		return &dto.CreateAuditLogResp{
			OK:         false,
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.CreateAuditLogResp{
		OK:         true,
		StatusCode: code.Success,
		StatusMsg:  "audit log created successfully",
	}, nil
}

// GetAuditLog 获取审计日志详情
func (s *AuditAppService) GetAuditLog(ctx context.Context, req *dto.GetAuditLogReq) (*dto.GetAuditLogResp, error) {
	auditLog, err := s.auditRepo.GetByID(ctx, req.ID)
	if err != nil {
		return &dto.GetAuditLogResp{
			AuditLog:   nil,
			StatusCode: code.AuditNotExist,
			StatusMsg:  code.AuditNotExistMsg,
		}, nil
	}

	return &dto.GetAuditLogResp{
		AuditLog:   s.convertToDTO(auditLog),
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// ListAuditLogs 查询审计日志列表
func (s *AuditAppService) ListAuditLogs(ctx context.Context, req *dto.ListAuditLogsReq) (*dto.ListAuditLogsResp, error) {
	var auditLogs []*entity.AuditLog
	var total int64
	var err error

	// 根据不同条件查询
	if req.TargetTable != "" && req.TargetID > 0 {
		auditLogs, total, err = s.auditRepo.ListByTarget(ctx, req.TargetTable, req.TargetID, req.Page, req.PageSize)
	} else if req.ActionType != "" {
		auditLogs, total, err = s.auditRepo.ListByActionType(ctx, req.ActionType, req.Page, req.PageSize)
	} else if req.UserID > 0 {
		auditLogs, total, err = s.auditRepo.ListByUserID(ctx, req.UserID, req.Page, req.PageSize)
	} else {
		// 默认查询所有
		auditLogs, total, err = s.auditRepo.ListByUserID(ctx, 0, req.Page, req.PageSize)
	}

	if err != nil {
		return &dto.ListAuditLogsResp{
			AuditLogs:  nil,
			TotalCount: 0,
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	items := make([]*dto.AuditLogDTO, 0, len(auditLogs))
	for _, log := range auditLogs {
		items = append(items, s.convertToDTO(log))
	}

	return &dto.ListAuditLogsResp{
		AuditLogs:  items,
		TotalCount: total,
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// GetByTraceID 根据TraceID查询
func (s *AuditAppService) GetByTraceID(ctx context.Context, traceID string) (*dto.GetAuditLogResp, error) {
	auditLog, err := s.auditRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return &dto.GetAuditLogResp{
			AuditLog:   nil,
			StatusCode: code.AuditNotExist,
			StatusMsg:  code.AuditNotExistMsg,
		}, nil
	}

	return &dto.GetAuditLogResp{
		AuditLog:   s.convertToDTO(auditLog),
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// 转换方法
func (s *AuditAppService) convertToDTO(log *entity.AuditLog) *dto.AuditLogDTO {
	return &dto.AuditLogDTO{
		ID:          log.ID,
		UserID:      log.UserID,
		ActionType:  log.ActionType,
		ActionDesc:  log.ActionDesc,
		OldData:     log.OldData,
		NewData:     log.NewData,
		ServiceName: log.ServiceName,
		TargetTable: log.TargetTable,
		TargetID:    log.TargetID,
		ClientIP:    log.ClientIP,
		TraceID:     log.TraceID,
		SpanID:      log.SpanID,
		CreatedAt:   log.CreatedAt.Unix(),
	}
}

func init() {
	_ = code.ServerError // 引入code包
}
