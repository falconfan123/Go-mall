package aggregate

import (
	"github.com/falconfan123/Go-mall/services/product/internal/domain/valueobject"
	"time"
)

// Product 商品聚合根，核心领域模型
type Product struct {
	ID           int64                  // 商品ID
	Name         string                 // 商品名称
	Description  string                 // 商品描述
	Picture      string                 // 商品主图
	Price        valueobject.Price      // 商品价格（分）
	Stock        valueobject.Stock      // 商品库存
	Sold         int64                  // 销量
	Categories   []valueobject.Category // 商品分类
	CreatedAt    time.Time              // 创建时间
	UpdatedAt    time.Time              // 更新时间
	ThumbnailURL string                 // 缩略图URL
}

// NewProduct 创建新商品
func NewProduct(
	name string,
	description string,
	picture string,
	price valueobject.Price,
	stock valueobject.Stock,
	categories []valueobject.Category,
	thumbnailURL string,
) *Product {
	now := time.Now()
	return &Product{
		Name:         name,
		Description:  description,
		Picture:      picture,
		Price:        price,
		Stock:        stock,
		Sold:         0,
		Categories:   categories,
		CreatedAt:    now,
		UpdatedAt:    now,
		ThumbnailURL: thumbnailURL,
	}
}

// UpdateInfo 更新商品基本信息
func (p *Product) UpdateInfo(
	name string,
	description string,
	picture string,
	price valueobject.Price,
	categories []valueobject.Category,
	thumbnailURL string,
) {
	p.Name = name
	p.Description = description
	p.Picture = picture
	p.Price = price
	p.Categories = categories
	p.ThumbnailURL = thumbnailURL
	p.UpdatedAt = time.Now()
}

// AdjustStock 调整库存（正数为增加，负数为减少）
func (p *Product) AdjustStock(quantity int64) error {
	newStock, err := p.Stock.Adjust(quantity)
	if err != nil {
		return err
	}
	p.Stock = newStock
	p.UpdatedAt = time.Now()
	return nil
}

// IncreaseSold 增加销量
func (p *Product) IncreaseSold(quantity int64) {
	p.Sold += quantity
	p.UpdatedAt = time.Now()
}

// IsOnSale 判断商品是否在售（库存>0）
func (p *Product) IsOnSale() bool {
	return p.Stock.Value() > 0
}

// HasCategory 判断商品是否属于指定分类
func (p *Product) HasCategory(category valueobject.Category) bool {
	for _, c := range p.Categories {
		if c == category {
			return true
		}
	}
	return false
}
