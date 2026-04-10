package user_address

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAddressesModel = (*customUserAddressesModel)(nil)

type (
	// UserAddressesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAddressesModel.
	UserAddressesModel interface {
		userAddressesModel
		GetUserAddressExistsByIdAndUserId(ctx context.Context, addressId uint64, userId int32) (bool, error)
		WithSession(session sqlx.Session) UserAddressesModel

		FindAllByUserId(ctx context.Context, userId int32) ([]*UserAddresses, error)
		DeleteByAddressIdandUserId(ctx context.Context, addressId uint64, userId int32) error
		InsertWithSession(ctx context.Context, session sqlx.Session, data *UserAddresses) (sql.Result, error)
		GetUserAddressbyIdAndUserId(ctx context.Context, addressId uint64, userId int32) (*UserAddresses, error)
		UpdateWithSession(ctx context.Context, session sqlx.Session, data *UserAddresses) (sql.Result, error)

		BatchUpdateDeFaultWithSession(ctx context.Context, session sqlx.Session, data []*UserAddresses) error
	}

	customUserAddressesModel struct {
		*defaultUserAddressesModel
	}
)

// NewUserAddressesModel returns a model for the database table.
func NewUserAddressesModel(conn sqlx.SqlConn, c cache.CacheConf) UserAddressesModel {
	return &customUserAddressesModel{
		defaultUserAddressesModel: newUserAddressesModel(conn),
	}
}

// WithSession method
func (m *customUserAddressesModel) WithSession(session sqlx.Session) UserAddressesModel {
	return NewUserAddressesModel(
		sqlx.NewSqlConnFromSession(session),
		cache.CacheConf{},
	)
}

func (m *customUserAddressesModel) FindAllByUserId(ctx context.Context, userId int32) ([]*UserAddresses, error) {
	var resp []*UserAddresses
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", userAddressesRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)

	switch {
	case err == nil:
		return resp, nil
	case err == sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserAddressesModel) DeleteByAddressIdandUserId(ctx context.Context, addressId uint64, userId int32) error {
	query := fmt.Sprintf("delete from %s where address_id = $1 and user_id = $2", m.table)
	_, err := m.conn.ExecCtx(ctx, query, addressId, userId)
	return err
}

func (m *customUserAddressesModel) GetUserAddressExistsByIdAndUserId(ctx context.Context, addressId uint64, userId int32) (bool, error) {
	var exists bool
	query := fmt.Sprintf("select exists(select 1 from %s where address_id = $1 and user_id = $2)", m.table)
	err := m.conn.QueryRowCtx(ctx, &exists, query, addressId, userId)

	switch err {
	case nil:
		return exists, nil
	case sqlx.ErrNotFound:
		return false, ErrNotFound
	default:
		return false, err
	}
}

func (m *customUserAddressesModel) GetUserAddressbyIdAndUserId(ctx context.Context, addressId uint64, userId int32) (*UserAddresses, error) {
	var resp UserAddresses
	query := fmt.Sprintf("select %s from %s where address_id = $1 and user_id = $2", userAddressesRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, addressId, userId)

	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserAddressesModel) BatchUpdateDeFaultWithSession(ctx context.Context, session sqlx.Session, data []*UserAddresses) error {
	for _, userAddress := range data {
		query := fmt.Sprintf("update %s set is_default = false where user_id = $1", m.table)
		_, err := session.ExecCtx(ctx, query, userAddress.UserId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *customUserAddressesModel) InsertWithSession(ctx context.Context, session sqlx.Session, data *UserAddresses) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values ($1, $2, $3, $4, $5, $6, $7)", m.table, userAddressesRowsExpectAutoSet)
	result, err := session.ExecCtx(ctx, query, data.UserId, data.DetailedAddress, data.City, data.Province, data.IsDefault, data.RecipientName, data.PhoneNumber)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *customUserAddressesModel) UpdateWithSession(ctx context.Context, session sqlx.Session, data *UserAddresses) (sql.Result, error) {
	query := fmt.Sprintf("update %s set %s where address_id = $1", m.table, userAddressesRowsWithPlaceHolder)
	result, err := session.ExecCtx(ctx, query, data.UserId, data.DetailedAddress, data.City, data.Province, data.IsDefault, data.RecipientName, data.PhoneNumber, data.AddressId)
	if err != nil {
		return nil, err
	}
	return result, nil
}
