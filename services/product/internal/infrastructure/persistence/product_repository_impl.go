package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/falconfan123/Go-mall/dal/model/products/product"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/product/internal/domain/valueobject"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ProductRepositoryImpl 商品仓储实现
type ProductRepositoryImpl struct {
	productModel product.ProductsModel
	conn         sqlx.SqlConn
}

// NewProductRepositoryImpl 创建商品仓储实现
func NewProductRepositoryImpl(conn sqlx.SqlConn) repository.ProductRepository {
	return &ProductRepositoryImpl{
		productModel: product.NewProductsModel(conn),
		conn:         conn,
	}
}

// Save 保存商品（新建或更新）
func (r *ProductRepositoryImpl) Save(ctx context.Context, productAgg *aggregate.Product) error {
	// 转换领域模型到数据模型
	productData := &product.Products{
		Id:          productAgg.ID,
		Name:        productAgg.Name,
		Description: sql.NullString{String: productAgg.Description, Valid: productAgg.Description != ""},
		Picture:     sql.NullString{String: productAgg.Picture, Valid: productAgg.Picture != ""},
		Price:       productAgg.Price.Value(),
		Stock:       productAgg.Stock.Value(),
		CreatedAt:   productAgg.CreatedAt,
		UpdatedAt:   productAgg.UpdatedAt,
	}

	if productAgg.ID == 0 {
		// 新建
		result, err := r.productModel.Insert(ctx, productData)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		productAgg.ID = id
		return nil
	}

	// 更新
	return r.productModel.Update(ctx, productData)
}

// GetByID 根据ID查询商品
func (r *ProductRepositoryImpl) GetByID(ctx context.Context, id int64) (*aggregate.Product, error) {
	productData, err := r.productModel.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}

	// 转换数据模型到领域模型
	price, err := valueobject.NewPrice(productData.Price)
	if err != nil {
		return nil, err
	}

	stock, err := valueobject.NewStock(productData.Stock)
	if err != nil {
		return nil, err
	}

	// 这里简化分类处理，实际应该查询关联的分类
	categories := make([]valueobject.Category, 0)

	return &aggregate.Product{
		ID:           productData.Id,
		Name:         productData.Name,
		Description:  productData.Description.String,
		Picture:      productData.Picture.String,
		Price:        price,
		Stock:        stock,
		Sold:         0, // 从其他表获取
		Categories:   categories,
		CreatedAt:    productData.CreatedAt,
		UpdatedAt:    productData.UpdatedAt,
		ThumbnailURL: "", // 从其他表获取
	}, nil
}

// List 查询商品列表
func (r *ProductRepositoryImpl) List(ctx context.Context, page, pageSize int, categoryID *int64, keyword *string) ([]*aggregate.Product, int64, error) {
	offset := (page - 1) * pageSize
	var products []*product.Products
	var total int64

	// 构建查询条件
	where := "1=1"
	args := make([]interface{}, 0)

	if keyword != nil && *keyword != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+*keyword+"%")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", "products", where)
	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?, ?", "id, name, description, picture, price, stock, created_at, updated_at", "products", where)
	args = append(args, offset, pageSize)
	err = r.conn.QueryRowsCtx(ctx, &products, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// 转换为领域模型
	productAggs := make([]*aggregate.Product, 0, len(products))
	for _, p := range products {
		price, err := valueobject.NewPrice(p.Price)
		if err != nil {
			continue
		}
		stock, err := valueobject.NewStock(p.Stock)
		if err != nil {
			continue
		}

		productAggs = append(productAggs, &aggregate.Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description.String,
			Picture:     p.Picture.String,
			Price:       price,
			Stock:       stock,
			Sold:        0,
			Categories:  []valueobject.Category{},
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	return productAggs, total, nil
}

// Delete 删除商品
func (r *ProductRepositoryImpl) Delete(ctx context.Context, id int64) error {
	return r.productModel.Delete(ctx, id)
}

// BatchGetByIDs 批量查询商品
func (r *ProductRepositoryImpl) BatchGetByIDs(ctx context.Context, ids []int64) ([]*aggregate.Product, error) {
	if len(ids) == 0 {
		return []*aggregate.Product{}, nil
	}

	// 构建IN查询
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id IN (%s)", "id, name, description, picture, price, stock, created_at, updated_at", "products", strings.Join(placeholders, ","))
	var products []*product.Products
	err := r.conn.QueryRowsCtx(ctx, &products, query, args...)
	if err != nil {
		return nil, err
	}

	// 转换为领域模型
	productAggs := make([]*aggregate.Product, 0, len(products))
	for _, p := range products {
		price, err := valueobject.NewPrice(p.Price)
		if err != nil {
			continue
		}
		stock, err := valueobject.NewStock(p.Stock)
		if err != nil {
			continue
		}

		productAggs = append(productAggs, &aggregate.Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description.String,
			Picture:     p.Picture.String,
			Price:       price,
			Stock:       stock,
			Sold:        0,
			Categories:  []valueobject.Category{},
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	return productAggs, nil
}

// DecreaseStock 扣减库存（原子操作）
func (r *ProductRepositoryImpl) DecreaseStock(ctx context.Context, productID int64, quantity int64) error {
	query := "UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?"
	result, err := r.conn.ExecCtx(ctx, query, quantity, productID, quantity)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return valueobject.ErrInsufficientStock
	}
	return nil
}

// IncreaseStock 增加库存（原子操作）
func (r *ProductRepositoryImpl) IncreaseStock(ctx context.Context, productID int64, quantity int64) error {
	query := "UPDATE products SET stock = stock + ? WHERE id = ?"
	_, err := r.conn.ExecCtx(ctx, query, quantity, productID)
	return err
}
