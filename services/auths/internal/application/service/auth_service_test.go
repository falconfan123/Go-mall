package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/auths/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/auths/internal/domain/repository"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationChecksClientIPAndLogoutTime(t *testing.T) {
	t.Parallel()

	logoutAt := time.Now()
	repo := &stubTokenRepository{
		claims: &repository.TokenClaims{
			UserID:   7,
			ClientIP: "127.0.0.1",
			IssuedAt: logoutAt.Add(-time.Second),
		},
		logoutTime: &logoutAt,
	}
	service := NewAuthAppService(repo)

	resp, err := service.Authentication(context.Background(), &dto.AuthReq{
		Token:    "token",
		ClientIP: "127.0.0.1",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.AuthExpiredByLogout, resp.StatusCode)
}

func TestGenerateTokenAndRenewToken(t *testing.T) {
	t.Parallel()

	repo := &stubTokenRepository{
		generatedToken: "new-token",
		claims: &repository.TokenClaims{
			UserID:   9,
			ClientIP: "127.0.0.1",
		},
	}
	service := NewAuthAppService(repo)

	generated, err := service.GenerateToken(context.Background(), &dto.GenerateTokenReq{
		UserID:   9,
		ClientIP: "127.0.0.1",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, generated.StatusCode)
	require.Equal(t, int64(7*24*3600), generated.ExpiresIn)

	renewed, err := service.RenewToken(context.Background(), &dto.RenewTokenReq{
		Token:    "old-token",
		ClientIP: "127.0.0.1",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, renewed.StatusCode)
	require.Equal(t, "new-token", renewed.Token)
}

func TestValidateTokenMapsRepositoryErrors(t *testing.T) {
	t.Parallel()

	service := NewAuthAppService(&stubTokenRepository{
		validateErr: repository.ErrTokenExpired,
	})
	resp, err := service.ValidateToken(context.Background(), "token", "127.0.0.1")
	require.NoError(t, err)
	require.EqualValues(t, code.AuthExpired, resp.StatusCode)

	service = NewAuthAppService(&stubTokenRepository{
		validateErr: errors.New("boom"),
	})
	_, err = service.ValidateToken(context.Background(), "token", "127.0.0.1")
	require.Error(t, err)
}

type stubTokenRepository struct {
	generatedToken string
	generateErr    error
	claims         *repository.TokenClaims
	parseErr       error
	validateErr    error
	logoutTime     *time.Time
	logoutErr      error
}

func (s *stubTokenRepository) GenerateToken(context.Context, int64, string, time.Duration) (string, error) {
	return s.generatedToken, s.generateErr
}

func (s *stubTokenRepository) ParseToken(context.Context, string) (*repository.TokenClaims, error) {
	return s.claims, s.parseErr
}

func (s *stubTokenRepository) ValidateToken(context.Context, string, string) (*repository.TokenClaims, error) {
	return s.claims, s.validateErr
}

func (s *stubTokenRepository) GetLogoutTime(context.Context, int64) (*time.Time, error) {
	return s.logoutTime, s.logoutErr
}

func (s *stubTokenRepository) SetLogoutTime(context.Context, int64, time.Time) error {
	return nil
}

var _ repository.TokenRepository = (*stubTokenRepository)(nil)
