package service

import (
	"context"
	"errors"
	"testing"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/audit/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/audit/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/audit/internal/domain/repository"
	"github.com/stretchr/testify/require"
)

func TestAuditAppServiceCreateAndGet(t *testing.T) {
	t.Parallel()

	repo := &stubAuditRepository{
		auditLog: &entity.AuditLog{
			ID:          1,
			UserID:      2,
			ActionType:  "update",
			ActionDesc:  "updated product",
			ServiceName: "product",
			TargetTable: "products",
			TargetID:    10,
			ClientIP:    "127.0.0.1",
		},
	}
	service := NewAuditAppService(repo)

	createResp, err := service.CreateAuditLog(context.Background(), &dto.CreateAuditLogReq{
		UserID:      2,
		ActionType:  "update",
		ServiceName: "product",
		TargetTable: "products",
		TargetID:    10,
		ClientIP:    "127.0.0.1",
	})
	require.NoError(t, err)
	require.True(t, createResp.OK)
	require.NotNil(t, repo.savedAuditLog)

	getResp, err := service.GetAuditLog(context.Background(), &dto.GetAuditLogReq{ID: 1})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, getResp.StatusCode)
	require.EqualValues(t, 2, getResp.AuditLog.UserID)
}

func TestAuditAppServiceListAndHandleRepositoryError(t *testing.T) {
	t.Parallel()

	service := NewAuditAppService(&stubAuditRepository{
		listLogs: []*entity.AuditLog{
			{ID: 1, ActionType: "create", ServiceName: "users", TargetTable: "users"},
		},
		total: 1,
	})
	resp, err := service.ListAuditLogs(context.Background(), &dto.ListAuditLogsReq{
		ActionType: "create",
		Page:       1,
		PageSize:   10,
	})
	require.NoError(t, err)
	require.Len(t, resp.AuditLogs, 1)
	require.EqualValues(t, 1, resp.TotalCount)

	service = NewAuditAppService(&stubAuditRepository{
		getErr: errors.New("missing"),
	})
	getResp, err := service.GetAuditLog(context.Background(), &dto.GetAuditLogReq{ID: 99})
	require.NoError(t, err)
	require.EqualValues(t, code.AuditNotExist, getResp.StatusCode)
}

type stubAuditRepository struct {
	auditLog      *entity.AuditLog
	getErr        error
	saveErr       error
	listLogs      []*entity.AuditLog
	total         int64
	listErr       error
	savedAuditLog *entity.AuditLog
}

func (s *stubAuditRepository) GetByID(context.Context, uint64) (*entity.AuditLog, error) {
	return s.auditLog, s.getErr
}

func (s *stubAuditRepository) GetByTraceID(context.Context, string) (*entity.AuditLog, error) {
	return s.auditLog, s.getErr
}

func (s *stubAuditRepository) Save(_ context.Context, auditLog *entity.AuditLog) error {
	s.savedAuditLog = auditLog
	return s.saveErr
}

func (s *stubAuditRepository) Delete(context.Context, uint64) error {
	return nil
}

func (s *stubAuditRepository) ListByUserID(context.Context, uint64, int, int) ([]*entity.AuditLog, int64, error) {
	return s.listLogs, s.total, s.listErr
}

func (s *stubAuditRepository) ListByTarget(context.Context, string, uint64, int, int) ([]*entity.AuditLog, int64, error) {
	return s.listLogs, s.total, s.listErr
}

func (s *stubAuditRepository) ListByActionType(context.Context, string, int, int) ([]*entity.AuditLog, int64, error) {
	return s.listLogs, s.total, s.listErr
}

var _ repository.AuditRepository = (*stubAuditRepository)(nil)
