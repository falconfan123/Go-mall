package product

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ ProductsModel = (*CustomProductsModel)(nil)

type (
	// ProductsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductsModel.
	ProductsModel interface {
		productsModel
		WithSession(session sqlx.Session) ProductsModel
		FindPage(ctx context.Context, offset, limit int) ([]*Products, error)
		Count(ctx context.Context) (int64, error)
		FindProductIsExist(ctx context.Context, productID int64) (bool, error)
		QueryAllProducts(ctx context.Context) ([]*Products, error)
		GetProductByIDs(ctx context.Context, productIDs []string) ([]*Products, error)
		FindListByCursor(ctx context.Context, cursor int64, limit int64) ([]*Products, error)
	}

	CustomProductsModel struct {
		*defaultProductsModel
	}
)

func (m *CustomProductsModel) FindListByCursor(ctx context.Context, cursor int64, limit int64) ([]*Products, error) {
	var query string
	var args []interface{}
	if cursor <= 0 {
		query = fmt.Sprintf("SELECT * FROM %s ORDER BY id DESC LIMIT $1", m.table)
		args = append(args, limit)
	} else {
		query = fmt.Sprintf("SELECT * FROM %s WHERE id < $1 ORDER BY id DESC LIMIT $2", m.table)
		args = append(args, cursor, limit)
	}

	var products []*Products
	err := m.conn.QueryRowsCtx(ctx, &products, query, args...)
	return products, err
}

func (m *CustomProductsModel) GetProductByIDs(ctx context.Context, productIDs []string) ([]*Products, error) {
	if len(productIDs) == 0 {
		return make([]*Products, 0), nil
	}
	ids := make([]string, len(productIDs))
	for i := range productIDs {
		ids[i] = fmt.Sprintf("$%d", i+1)
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE id IN (%s)", m.table, strings.Join(ids, ","))
	products := make([]*Products, 0)

	err := m.conn.QueryRowsCtx(ctx, &products, query, strings.Join(productIDs, ","))
	return products, err
}

func (m *CustomProductsModel) QueryAllProducts(ctx context.Context) ([]*Products, error) {
	query := fmt.Sprintf("SELECT * FROM %s", m.table)
	products := make([]*Products, 0)
	err := m.conn.QueryRowsCtx(ctx, &products, query)
	return products, err
}

// NewProductsModel returns a model for the database table.
func NewProductsModel(conn sqlx.SqlConn) ProductsModel {
	return &CustomProductsModel{
		defaultProductsModel: newProductsModel(conn),
	}
}

func (m *CustomProductsModel) WithSession(session sqlx.Session) ProductsModel {
	return NewProductsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultProductsModel) FindPage(ctx context.Context, offset, limit int) ([]*Products, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", m.table)
	var products []*Products
	err := m.conn.QueryRowsCtx(ctx, &products, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return products, nil
}
func (m *defaultProductsModel) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (m *defaultProductsModel) FindProductIsExist(ctx context.Context, productID int64) (bool, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id=$1", m.table)

	err := m.conn.QueryRowCtx(ctx, &count, query, productID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
