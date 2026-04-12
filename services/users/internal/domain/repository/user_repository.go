package repository

import (
	"context"
	"errors"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/valueobject"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Save 保存用户
	Save(ctx context.Context, user *aggregate.User) (int64, error)
	// Update 更新用户
	Update(ctx context.Context, user *aggregate.User) error
	// FindByID 根据ID查找用户
	FindByID(ctx context.Context, userID int64) (*aggregate.User, error)
	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email *valueobject.Email) (*aggregate.User, error)
	// FindByUsernameOrEmail 根据用户名或邮箱查找用户
	FindByUsernameOrEmail(ctx context.Context, account string) (*aggregate.User, error)
	// ExistsByEmail 判断邮箱是否存在
	ExistsByEmail(ctx context.Context, email *valueobject.Email) (bool, error)
	// ExistsByUsername 判断用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	// UpdateLogoutTime 更新登出时间
	UpdateLogoutTime(ctx context.Context, userID int64, logoutTime time.Time) error
	// SaveAddress 保存地址
	SaveAddress(ctx context.Context, userID int64, address *entity.Address) (int64, error)
	// UpdateAddress 更新地址
	UpdateAddress(ctx context.Context, userID int64, address *entity.Address) error
	// DeleteAddress 删除地址
	DeleteAddress(ctx context.Context, userID int64, addressID int64) error
	// FindAddressesByUserID 查询用户所有地址
	FindAddressesByUserID(ctx context.Context, userID int64) ([]*entity.Address, error)
}
