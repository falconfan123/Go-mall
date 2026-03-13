package persistence

import (
	"context"
	"database/sql"
	"errors"
	daluser "github.com/falconfan123/Go-mall/dal/model/user"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/users/internal/domain/valueobject"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepositoryImpl 用户仓储实现
type UserRepositoryImpl struct {
	userModel daluser.UsersModel
}

// NewUserRepositoryImpl 创建用户仓储实现
func NewUserRepositoryImpl(userModel daluser.UsersModel) repository.UserRepository {
	return &UserRepositoryImpl{
		userModel: userModel,
	}
}

// Save 保存用户
func (r *UserRepositoryImpl) Save(ctx context.Context, user *aggregate.User) (int64, error) {
	u := &daluser.Users{
		Email:        sql.NullString{String: user.Email.Value(), Valid: true},
		PasswordHash: sql.NullString{String: user.PasswordHash.Value(), Valid: true},
		Username:     sql.NullString{String: user.Username, Valid: true},
		AvatarUrl:    sql.NullString{String: user.Avatar, Valid: user.Avatar != ""},
		LoginAt:      sql.NullTime{Time: user.LastLoginTime, Valid: !user.LastLoginTime.IsZero()},
	}

	res, err := r.userModel.Insert(ctx, u)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Update 更新用户
func (r *UserRepositoryImpl) Update(ctx context.Context, user *aggregate.User) error {
	u := &daluser.Users{
		UserId:       user.ID,
		Email:        sql.NullString{String: user.Email.Value(), Valid: true},
		PasswordHash: sql.NullString{String: user.PasswordHash.Value(), Valid: true},
		Username:     sql.NullString{String: user.Username, Valid: true},
		AvatarUrl:    sql.NullString{String: user.Avatar, Valid: user.Avatar != ""},
		LoginAt:      sql.NullTime{Time: user.LastLoginTime, Valid: !user.LastLoginTime.IsZero()},
	}
	_, err := r.userModel.Update(ctx, u)
	return err
}

// FindByID 根据ID查找用户
func (r *UserRepositoryImpl) FindByID(ctx context.Context, userID int64) (*aggregate.User, error) {
	u, err := r.userModel.FindOne(ctx, userID)
	if err != nil {
		if err == daluser.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return r.toDomainUser(u), nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email *valueobject.Email) (*aggregate.User, error) {
	u, err := r.userModel.FindOneByEmail(ctx, sql.NullString{String: email.Value(), Valid: true})
	if err != nil {
		if err == daluser.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return r.toDomainUser(u), nil
}

// FindByUsernameOrEmail 根据用户名或邮箱查找用户
func (r *UserRepositoryImpl) FindByUsernameOrEmail(ctx context.Context, account string) (*aggregate.User, error) {
	u, err := r.userModel.FindOneByEmailOrUsername(ctx, account)
	if err != nil {
		if err == daluser.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return r.toDomainUser(u), nil
}

// ExistsByEmail 判断邮箱是否存在
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email *valueobject.Email) (bool, error) {
	_, err := r.userModel.FindOneByEmail(ctx, sql.NullString{String: email.Value(), Valid: true})
	if err != nil {
		if err == daluser.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistsByUsername 判断用户名是否存在
func (r *UserRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, err := r.userModel.FindOneByUsername(ctx, username)
	if err != nil {
		if err == daluser.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdateLogoutTime 更新登出时间
func (r *UserRepositoryImpl) UpdateLogoutTime(ctx context.Context, userID int64, logoutTime time.Time) error {
	return r.userModel.UpdateLogoutTime(ctx, userID, logoutTime)
}

// SaveAddress 保存地址
func (r *UserRepositoryImpl) SaveAddress(ctx context.Context, userID int64, address *entity.Address) (int64, error) {
	// TODO: 实现地址保存逻辑，需要先有地址的Model
	return 0, errors.New("not implemented")
}

// UpdateAddress 更新地址
func (r *UserRepositoryImpl) UpdateAddress(ctx context.Context, userID int64, address *entity.Address) error {
	// TODO: 实现地址更新逻辑
	return errors.New("not implemented")
}

// DeleteAddress 删除地址
func (r *UserRepositoryImpl) DeleteAddress(ctx context.Context, userID int64, addressID int64) error {
	// TODO: 实现地址删除逻辑
	return errors.New("not implemented")
}

// FindAddressesByUserID 查询用户所有地址
func (r *UserRepositoryImpl) FindAddressesByUserID(ctx context.Context, userID int64) ([]*entity.Address, error) {
	// TODO: 实现地址查询逻辑
	return nil, errors.New("not implemented")
}

// 转换数据库模型到领域模型
func (r *UserRepositoryImpl) toDomainUser(u *daluser.Users) *aggregate.User {
	email, _ := valueobject.NewEmail(u.Email.String)
	passwordHash := valueobject.NewPasswordHashFromHash(u.PasswordHash.String)

	domainUser := &aggregate.User{
		ID:            u.UserId,
		Email:         email,
		PasswordHash:  passwordHash,
		Username:      u.Username.String,
		Avatar:        u.AvatarUrl.String,
		CreateTime:    u.CreatedAt,
		UpdateTime:    u.UpdatedAt,
		LastLoginTime: u.LoginAt.Time,
		Addresses:     make([]*entity.Address, 0),
	}
	return domainUser
}
