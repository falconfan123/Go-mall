package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	authsclient "github.com/falconfan123/Go-mall/services/auths/authsclient"
	"github.com/falconfan123/Go-mall/services/users/internal/application/dto"
	appevent "github.com/falconfan123/Go-mall/services/users/internal/application/event"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/entity"
	domainevent "github.com/falconfan123/Go-mall/services/users/internal/domain/event"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestAuthAppServiceRegister(t *testing.T) {
	t.Parallel()

	repo := &stubUserRepository{
		saveID: 101,
	}
	publisher := &stubUserEventPublisher{}
	service := NewAuthAppService(repo, publisher, &stubAuthsClient{})

	resp, err := service.Register(context.Background(), &dto.RegisterRequest{
		Email:           "user@example.com",
		Password:        "secret-123",
		ConfirmPassword: "secret-123",
		IP:              "127.0.0.1",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, resp.StatusCode)
	require.EqualValues(t, 101, resp.UserID)
	require.NotEmpty(t, resp.ShortToken)
	require.NotEmpty(t, resp.LongToken)
	require.NotNil(t, repo.savedUser)
	require.Equal(t, "user@example.com", repo.savedUser.Username)

	require.Eventually(t, func() bool {
		return publisher.registeredCount() == 1
	}, time.Second, 10*time.Millisecond)
}

func TestAuthAppServiceRegisterRejectsDuplicateUsername(t *testing.T) {
	t.Parallel()

	service := NewAuthAppService(&stubUserRepository{
		existsByUsername: true,
	}, &stubUserEventPublisher{}, &stubAuthsClient{})

	resp, err := service.Register(context.Background(), &dto.RegisterRequest{
		Username:        "fan",
		Password:        "secret-123",
		ConfirmPassword: "secret-123",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.UserExistError, resp.StatusCode)
}

func TestAuthAppServiceLogin(t *testing.T) {
	t.Parallel()

	email, err := valueobject.NewEmail("user@example.com")
	require.NoError(t, err)
	user := aggregate.NewUser(email, valueobject.NewPasswordHash("secret-123"), "fan")
	user.ID = 7

	repo := &stubUserRepository{
		userByAccount: user,
	}
	publisher := &stubUserEventPublisher{}
	service := NewAuthAppService(repo, publisher, &stubAuthsClient{})

	resp, err := service.Login(context.Background(), &dto.LoginRequest{
		Username: "fan",
		Password: "secret-123",
		IP:       "127.0.0.1",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, resp.StatusCode)
	require.EqualValues(t, 7, repo.loginTimeUpdated)
	require.Nil(t, repo.updatedUser)

	require.Eventually(t, func() bool {
		return publisher.loginCount() == 1
	}, time.Second, 10*time.Millisecond)
}

func TestAuthAppServiceLoginFailsWhenLoginTimeUpdateFails(t *testing.T) {
	t.Parallel()

	email, err := valueobject.NewEmail("user@example.com")
	require.NoError(t, err)
	user := aggregate.NewUser(email, valueobject.NewPasswordHash("secret-123"), "fan")
	user.ID = 8

	service := NewAuthAppService(&stubUserRepository{
		userByAccount:  user,
		updateLoginErr: errors.New("write failed"),
	}, &stubUserEventPublisher{}, &stubAuthsClient{})

	resp, err := service.Login(context.Background(), &dto.LoginRequest{
		Username: "fan",
		Password: "secret-123",
		IP:       "127.0.0.1",
	})
	require.Error(t, err)
	require.EqualValues(t, code.ServerError, resp.StatusCode)
}

func TestAuthAppServiceLoginUserNotFound(t *testing.T) {
	t.Parallel()

	service := NewAuthAppService(&stubUserRepository{
		findByAccountErr: repository.ErrUserNotFound,
	}, &stubUserEventPublisher{}, &stubAuthsClient{})

	resp, err := service.Login(context.Background(), &dto.LoginRequest{
		Username: "fan",
		Password: "secret-123",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.UserNotExistError, resp.StatusCode)
}

func TestAuthAppServiceLogoutFailsWhenLogoutTimeUpdateFails(t *testing.T) {
	t.Parallel()

	service := NewAuthAppService(&stubUserRepositoryWithLogoutError{
		err: errors.New("write failed"),
	}, &stubUserEventPublisher{}, &stubAuthsClient{})

	resp, err := service.Logout(context.Background(), &dto.LogoutRequest{
		UserID:    9,
		LongToken: "long-token",
		IP:        "127.0.0.1",
	})
	require.Error(t, err)
	require.EqualValues(t, code.ServerError, resp.StatusCode)
}

type stubUserRepository struct {
	saveID            int64
	saveErr           error
	existsByEmail     bool
	existsByEmailErr  error
	existsByUsername  bool
	existsByUserErr   error
	userByAccount     *aggregate.User
	findByAccountErr  error
	updateErr         error
	updateLoginErr    error
	savedUser         *aggregate.User
	updatedUser       *aggregate.User
	loginTimeUpdated  int64
	logoutTimeUpdated int64
}

func (s *stubUserRepository) Save(_ context.Context, user *aggregate.User) (int64, error) {
	s.savedUser = user
	return s.saveID, s.saveErr
}

func (s *stubUserRepository) Update(_ context.Context, user *aggregate.User) error {
	s.updatedUser = user
	return s.updateErr
}

func (s *stubUserRepository) UpdateLoginTime(_ context.Context, userID int64, _ time.Time) error {
	s.loginTimeUpdated = userID
	return s.updateLoginErr
}

func (s *stubUserRepository) FindByID(context.Context, int64) (*aggregate.User, error) {
	return nil, repository.ErrUserNotFound
}

func (s *stubUserRepository) FindByEmail(context.Context, *valueobject.Email) (*aggregate.User, error) {
	return nil, repository.ErrUserNotFound
}

func (s *stubUserRepository) FindByUsernameOrEmail(context.Context, string) (*aggregate.User, error) {
	if s.findByAccountErr != nil {
		return nil, s.findByAccountErr
	}
	return s.userByAccount, nil
}

func (s *stubUserRepository) ExistsByEmail(context.Context, *valueobject.Email) (bool, error) {
	return s.existsByEmail, s.existsByEmailErr
}

func (s *stubUserRepository) ExistsByUsername(context.Context, string) (bool, error) {
	return s.existsByUsername, s.existsByUserErr
}

func (s *stubUserRepository) UpdateLogoutTime(context.Context, int64, time.Time) error {
	return nil
}

func (s *stubUserRepository) SaveAddress(context.Context, int64, *entity.Address) (int64, error) {
	return 0, errors.New("not implemented")
}

func (s *stubUserRepository) UpdateAddress(context.Context, int64, *entity.Address) error {
	return errors.New("not implemented")
}

func (s *stubUserRepository) DeleteAddress(context.Context, int64, int64) error {
	return errors.New("not implemented")
}

func (s *stubUserRepository) FindAddressesByUserID(context.Context, int64) ([]*entity.Address, error) {
	return nil, errors.New("not implemented")
}

type stubUserEventPublisher struct {
	registered atomic.Int32
	loggedIn   atomic.Int32
}

type stubAuthsClient struct{}

type stubUserRepositoryWithLogoutError struct {
	stubUserRepository
	err error
}

func (s *stubUserRepositoryWithLogoutError) UpdateLogoutTime(context.Context, int64, time.Time) error {
	return s.err
}

func (s *stubUserEventPublisher) PublishUserRegistered(*domainevent.UserRegisteredEvent) error {
	s.registered.Add(1)
	return nil
}

func (s *stubUserEventPublisher) PublishUserLoggedIn(*domainevent.UserLoggedInEvent) error {
	s.loggedIn.Add(1)
	return nil
}

func (s *stubAuthsClient) GenerateToken(context.Context, *authsclient.AuthGenReq, ...grpc.CallOption) (*authsclient.AuthGenRes, error) {
	return &authsclient.AuthGenRes{
		StatusCode:     code.Success,
		StatusMsg:      code.SuccessMsg,
		ShortToken:     "short-token",
		LongToken:      "long-token",
		ShortExpiresIn: time.Now().Add(24 * time.Hour).Unix(),
		LongExpiresIn:  time.Now().Add(30 * 24 * time.Hour).Unix(),
	}, nil
}

func (s *stubAuthsClient) RenewToken(context.Context, *authsclient.AuthRenewalReq, ...grpc.CallOption) (*authsclient.AuthRenewalRes, error) {
	return nil, errors.New("not implemented")
}

func (s *stubAuthsClient) ValidateToken(context.Context, *authsclient.AuthValidateReq, ...grpc.CallOption) (*authsclient.AuthValidateRes, error) {
	return nil, errors.New("not implemented")
}

func (s *stubAuthsClient) Logout(context.Context, *authsclient.LogoutReq, ...grpc.CallOption) (*authsclient.LogoutRes, error) {
	return &authsclient.LogoutRes{
		StatusCode: code.Success,
		StatusMsg:  code.SuccessMsg,
	}, nil
}

func (s *stubUserEventPublisher) PublishUserLoggedOut(*domainevent.UserLoggedOutEvent) error {
	return nil
}

func (s *stubUserEventPublisher) PublishUserInfoUpdated(*domainevent.UserInfoUpdatedEvent) error {
	return nil
}

func (s *stubUserEventPublisher) PublishAddressAdded(*domainevent.AddressAddedEvent) error {
	return nil
}

func (s *stubUserEventPublisher) PublishAddressUpdated(*domainevent.AddressUpdatedEvent) error {
	return nil
}

func (s *stubUserEventPublisher) PublishAddressDeleted(*domainevent.AddressDeletedEvent) error {
	return nil
}

func (s *stubUserEventPublisher) registeredCount() int {
	return int(s.registered.Load())
}

func (s *stubUserEventPublisher) loginCount() int {
	return int(s.loggedIn.Load())
}

var _ repository.UserRepository = (*stubUserRepository)(nil)
var _ appevent.EventPublisher = (*stubUserEventPublisher)(nil)
