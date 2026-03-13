package service

import (
	"context"
	"errors"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/utils/token"
	"github.com/falconfan123/Go-mall/services/users/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/users/internal/application/event"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/aggregate"
	domainevent "github.com/falconfan123/Go-mall/services/users/internal/domain/event"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/valueobject"
	"github.com/google/uuid"
	"time"
)

// AuthAppService 认证应用服务
type AuthAppService struct {
	userRepo       repository.UserRepository
	eventPublisher event.EventPublisher
	authConfig     *AuthConfig
}

// AuthConfig 认证配置
type AuthConfig struct {
	AccessExpire  int64  // 访问令牌有效期（秒）
	RefreshExpire int64  // 刷新令牌有效期（秒）
	Secret        string // JWT密钥
}

// NewAuthAppService 创建认证应用服务
func NewAuthAppService(
	userRepo repository.UserRepository,
	eventPublisher event.EventPublisher,
	authConfig *AuthConfig,
) *AuthAppService {
	return &AuthAppService{
		userRepo:       userRepo,
		eventPublisher: eventPublisher,
		authConfig:     authConfig,
	}
}

// Register 用户注册
func (s *AuthAppService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 1. 校验参数
	if req.Password != req.ConfirmPassword {
		return &dto.RegisterResponse{
			StatusCode: uint32(code.RePasswordError),
			StatusMsg:  code.RePasswordErrorMsg,
		}, nil
	}

	// 2. 处理用户名和邮箱
	// 支持纯用户名注册（不需要邮箱）
	var email *valueobject.Email
	var err error

	if req.Email != "" {
		// 如果提供了邮箱，验证邮箱格式
		email, err = valueobject.NewEmail(req.Email)
		if err != nil {
			return &dto.RegisterResponse{
				StatusCode: uint32(code.EmailFormatError),
				StatusMsg:  code.EmailFormatErrorMsg,
			}, nil
		}
	}

	// 3. 检查用户名和邮箱是否存在
	// 检查用户名是否存在
	if req.Username != "" {
		exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
		if err != nil {
			return &dto.RegisterResponse{
				StatusCode: uint32(code.ServerError),
				StatusMsg:  code.ServerErrorMsg,
			}, err
		}
		if exists {
			return &dto.RegisterResponse{
				StatusCode: uint32(code.UserExistError),
				StatusMsg:  code.UserExistErrorMsg,
			}, nil
		}
	}

	// 如果提供了邮箱，检查邮箱是否存在
	if email != nil {
		exists, err := s.userRepo.ExistsByEmail(ctx, email)
		if err != nil {
			return &dto.RegisterResponse{
				StatusCode: uint32(code.ServerError),
				StatusMsg:  code.ServerErrorMsg,
			}, err
		}
		if exists {
			return &dto.RegisterResponse{
				StatusCode: uint32(code.UserExistError),
				StatusMsg:  code.UserExistErrorMsg,
			}, nil
		}
	}

	// 4. 创建密码哈希
	passwordHash := valueobject.NewPasswordHash(req.Password)

	// 5. 创建用户聚合
	username := req.Username
	if username == "" {
		username = req.Email // 默认用户名为邮箱
	}
	user := aggregate.NewUser(email, passwordHash, username)

	// 6. 保存用户
	userID, err := s.userRepo.Save(ctx, user)
	if err != nil {
		return &dto.RegisterResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}
	user.ID = userID

	// 7. 生成令牌
	accessToken, refreshToken, err := s.generateTokens(userID, req.IP)
	if err != nil {
		return &dto.RegisterResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	// 8. 发布用户注册事件
	go func() {
		event := &domainevent.UserRegisteredEvent{
			BaseEvent: domainevent.BaseEvent{
				EventID:    uuid.NewString(),
				EventType:  "user_registered",
				OccurredAt: time.Now(),
			},
			UserID:   userID,
			Email:    email.Value(),
			Username: username,
			IP:       req.IP,
		}
		_ = s.eventPublisher.PublishUserRegistered(event)
	}()

	// 9. 返回响应
	return &dto.RegisterResponse{
		StatusCode:   0,
		StatusMsg:    "success",
		UserID:       uint32(userID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login 用户登录
func (s *AuthAppService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1. 确定登录账号（优先使用用户名，其次使用邮箱）
	account := req.Username
	if account == "" {
		account = req.Email
	}
	if account == "" {
		return &dto.LoginResponse{
			StatusCode: uint32(code.EmailFormatError),
			StatusMsg:  "email or username is required",
		}, nil
	}

	// 2. 查询用户（支持用户名或邮箱）
	user, err := s.userRepo.FindByUsernameOrEmail(ctx, account)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &dto.LoginResponse{
				StatusCode: uint32(code.UserNotExistError),
				StatusMsg:  code.UserNotExistErrorMsg,
			}, nil
		}
		return &dto.LoginResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	// 3. 验证密码
	if !user.VerifyPassword(req.Password) {
		return &dto.LoginResponse{
			StatusCode: uint32(code.LoginError),
			StatusMsg:  code.LoginErrorMsg,
		}, nil
	}

	// 4. 记录登录信息
	user.RecordLogin(req.IP)
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		// 记录日志，但不影响登录流程
	}

	// 5. 生成令牌
	accessToken, refreshToken, err := s.generateTokens(user.ID, req.IP)
	if err != nil {
		return &dto.LoginResponse{
			StatusCode: uint32(code.ServerError),
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	// 6. 发布登录事件
	go func() {
		event := &domainevent.UserLoggedInEvent{
			BaseEvent: domainevent.BaseEvent{
				EventID:    uuid.NewString(),
				EventType:  "user_logged_in",
				OccurredAt: time.Now(),
			},
			UserID: user.ID,
			Email:  user.Email.Value(),
			IP:     req.IP,
		}
		_ = s.eventPublisher.PublishUserLoggedIn(event)
	}()

	// 7. 返回响应
	return &dto.LoginResponse{
		StatusCode:   0,
		StatusMsg:    "success",
		UserID:       uint32(user.ID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout 用户登出
func (s *AuthAppService) Logout(ctx context.Context, req *dto.LogoutRequest) (*dto.LogoutResponse, error) {
	// 1. 更新登出时间
	err := s.userRepo.UpdateLogoutTime(ctx, int64(req.UserID), time.Now())
	if err != nil {
		// 记录日志，但返回成功
	}

	// 2. 发布登出事件
	go func() {
		event := &domainevent.UserLoggedOutEvent{
			BaseEvent: domainevent.BaseEvent{
				EventID:    uuid.NewString(),
				EventType:  "user_logged_out",
				OccurredAt: time.Now(),
			},
			UserID: int64(req.UserID),
			IP:     req.IP,
		}
		_ = s.eventPublisher.PublishUserLoggedOut(event)
	}()

	// 3. 返回响应
	return &dto.LogoutResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}, nil
}

// 辅助方法：生成访问令牌和刷新令牌
func (s *AuthAppService) generateTokens(userID int64, ip string) (accessToken string, refreshToken string, err error) {
	// 生成访问令牌
	accessToken, err = token.GenerateJWT(
		uint32(userID),
		"", // role
		ip,
		time.Duration(s.authConfig.AccessExpire)*time.Second,
	)
	if err != nil {
		return "", "", err
	}

	// 生成刷新令牌
	refreshToken, err = token.GenerateJWT(
		uint32(userID),
		"",
		ip,
		time.Duration(s.authConfig.RefreshExpire)*time.Second,
	)
	if err != nil {
		return accessToken, "", err // 访问令牌生成成功的话，尽量返回
	}

	return accessToken, refreshToken, nil
}
