package svc

import (
	"github.com/falconfan123/Go-mall/dal/model/user"
	"github.com/falconfan123/Go-mall/dal/model/user_address"
	"github.com/falconfan123/Go-mall/services/users/internal/application/event"
	"github.com/falconfan123/Go-mall/services/users/internal/application/service"
	"github.com/falconfan123/Go-mall/services/users/internal/config"
	domainevent "github.com/falconfan123/Go-mall/services/users/internal/domain/event"
	"github.com/falconfan123/Go-mall/services/users/internal/infrastructure/persistence"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	UsersModel         user.UsersModel
	UserAddressesModel user_address.UserAddressesModel
	AuthAppService     *service.AuthAppService
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化仓储
	userRepo := persistence.NewUserRepositoryImpl(user.NewUsersModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)))

	// 初始化事件发布器（暂时用空实现，后面可以替换为RabbitMQ实现）
	var eventPublisher event.EventPublisher = &NoopEventPublisher{}

	// 初始化应用服务
	authConfig := &service.AuthConfig{
		AccessExpire:  c.AuthConfig.AccessExpire,
		RefreshExpire: c.AuthConfig.AccessExpire * 2,
		Secret:        c.AuthConfig.AccessSecret,
	}
	authAppService := service.NewAuthAppService(userRepo, eventPublisher, authConfig)

	return &ServiceContext{
		Config:             c,
		UsersModel:         user.NewUsersModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		UserAddressesModel: user_address.NewUserAddressesModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource), c.Cache),
		AuthAppService:     authAppService,
	}
}

// NoopEventPublisher 空事件发布器，临时实现
type NoopEventPublisher struct{}

func (n *NoopEventPublisher) PublishUserRegistered(e *domainevent.UserRegisteredEvent) error {
	return nil
}
func (n *NoopEventPublisher) PublishUserLoggedIn(e *domainevent.UserLoggedInEvent) error { return nil }
func (n *NoopEventPublisher) PublishUserLoggedOut(e *domainevent.UserLoggedOutEvent) error {
	return nil
}
func (n *NoopEventPublisher) PublishUserInfoUpdated(e *domainevent.UserInfoUpdatedEvent) error {
	return nil
}
func (n *NoopEventPublisher) PublishAddressAdded(e *domainevent.AddressAddedEvent) error { return nil }
func (n *NoopEventPublisher) PublishAddressUpdated(e *domainevent.AddressUpdatedEvent) error {
	return nil
}
func (n *NoopEventPublisher) PublishAddressDeleted(e *domainevent.AddressDeletedEvent) error {
	return nil
}
