package product

import (
	product2 "github.com/falconfan123/Go-mall/dal/model/products/product"
)

// ProductDocument 是写入 products 索引的 ES 文档。
// 字段名为 snake_case，与 mapping.json / template.json 及查询逻辑（queryproductlogic.go）一致，
// 避免 Go 默认序列化的 PascalCase 字段导致查询按 created_at 等排序时报 search_phase_execution_exception。
type ProductDocument struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Picture     string `json:"picture"`
	Price       int64  `json:"price"`
	Stock       int64  `json:"stock"`
	Category    string `json:"category"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

const esTimeLayout = "2006-01-02 15:04:05"

// NewProductDocument 由 DB 模型构造 ES 文档（snake_case）。
// category 当前缺省为空串（分类索引完整性属后续工作，见 fix-product-search-index-readiness）。
func NewProductDocument(p *product2.Products) *ProductDocument {
	doc := &ProductDocument{
		Id:    p.Id,
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}
	if p.Description.Valid {
		doc.Description = p.Description.String
	}
	if p.Picture.Valid {
		doc.Picture = p.Picture.String
	}
	doc.CreatedAt = p.CreatedAt.Format(esTimeLayout)
	doc.UpdatedAt = p.UpdatedAt.Format(esTimeLayout)
	return doc
}
