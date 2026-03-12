package service

import (
	"context"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/auths/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/auths/internal/domain/repository"
)

// AuthAppService 认证应用服务
type AuthAppService struct {
	tokenRepo repository.TokenRepository
}

// NewAuthAppService 创建认证应用服务
func NewAuthAppService(tokenRepo repository.TokenRepository) *AuthAppService {
	return &AuthAppService{
		tokenRepo: tokenRepo,
	}
}

// Authentication 认证
func (s *AuthAppService) Authentication(ctx context.Context, req *dto.AuthReq) (*dto.AuthRes, error) {
	// 解析Token
	claims, err := s.tokenRepo.ParseToken(ctx, req.Token)
	if err != nil {
		return &dto.AuthRes{
			StatusCode: code.TokenValid,
			StatusMsg:  code.TokenInvalidMsg,
		}, nil
	}

	// 验证客户端IP
	if req.ClientIP == "" {
		return &dto.AuthRes{
			StatusCode: code.NotWithClientIP,
			StatusMsg:  code.NotWithClientIPMsg,
		}, nil
	}

	if req.ClientIP != claims.ClientIP {
		return &dto.AuthRes{
			StatusCode: code.AuthExpired,
			StatusMsg:  code.AuthExpiredMsg,
		}, nil
	}

	// 检查用户登出时间
	logoutTime, err := s.tokenRepo.GetLogoutTime(ctx, claims.UserID)
	if err != nil && err != repository.ErrLogoutTimeNotFound {
		return nil, err
	}

	if logoutTime != nil && claims.IssuedAt.Before(*logoutTime) {
		return &dto.AuthRes{
			StatusCode: code.AuthExpiredByLogout,
			StatusMsg:  code.AuthExpiredByLogoutMsg,
		}, nil
	}

	return &dto.AuthRes{
		UserID:     claims.UserID,
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// GenerateToken 生成Token
func (s *AuthAppService) GenerateToken(ctx context.Context, req *dto.GenerateTokenReq) (*dto.GenerateTokenResp, error) {
	expiry := time.Duration(req.ExpiresIn) * time.Second
	if req.ExpiresIn == 0 {
		expiry = 7 * 24 * time.Hour // 默认7天
	}

	token, err := s.tokenRepo.GenerateToken(ctx, req.UserID, req.ClientIP, expiry)
	if err != nil {
		return &dto.GenerateTokenResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.GenerateTokenResp{
		Token:      token,
		ExpiresIn:  int64(expiry.Seconds()),
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// RenewToken 续期Token
func (s *AuthAppService) RenewToken(ctx context.Context, req *dto.RenewTokenReq) (*dto.RenewTokenResp, error) {
	// 解析原Token
	claims, err := s.tokenRepo.ParseToken(ctx, req.Token)
	if err != nil {
		return &dto.RenewTokenResp{
			StatusCode: code.TokenValid,
			StatusMsg:  code.TokenInvalidMsg,
		}, nil
	}

	// 验证客户端IP
	if req.ClientIP != claims.ClientIP {
		return &dto.RenewTokenResp{
			StatusCode: code.AuthExpired,
			StatusMsg:  code.AuthExpiredMsg,
		}, nil
	}

	// 生成新Token
	newToken, err := s.tokenRepo.GenerateToken(ctx, claims.UserID, claims.ClientIP, 7*24*time.Hour)
	if err != nil {
		return &dto.RenewTokenResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.RenewTokenResp{
		Token:      newToken,
		ExpiresIn:  7 * 24 * 3600,
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// ValidateToken 验证Token
func (s *AuthAppService) ValidateToken(ctx context.Context, token string, clientIP string) (*dto.AuthRes, error) {
	claims, err := s.tokenRepo.ValidateToken(ctx, token, clientIP)
	if err != nil {
		if err == repository.ErrTokenExpired {
			return &dto.AuthRes{
				StatusCode: code.AuthExpired,
				StatusMsg:  code.AuthExpiredMsg,
			}, nil
		}
		if err == repository.ErrTokenInvalid {
			return &dto.AuthRes{
				StatusCode: code.TokenValid,
				StatusMsg:  code.TokenInvalidMsg,
			}, nil
		}
		return nil, err
	}

	return &dto.AuthRes{
		UserID:     claims.UserID,
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

func init() {
	_ = code.ServerError // 引入code包
}
