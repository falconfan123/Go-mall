package valueobject

import "errors"

// Category 商品分类值对象，不可变
type Category struct {
	ID   int64  // 分类ID
	Name string // 分类名称
}

var (
	ErrInvalidCategoryID   = errors.New("category ID cannot be negative")
	ErrInvalidCategoryName = errors.New("category name cannot be empty")
)

// NewCategory 创建分类值对象
func NewCategory(id int64, name string) (Category, error) {
	if id < 0 {
		return Category{}, ErrInvalidCategoryID
	}
	if name == "" {
		return Category{}, ErrInvalidCategoryName
	}
	return Category{
		ID:   id,
		Name: name,
	}, nil
}

// Equals 判断两个分类是否相等
func (c Category) Equals(other Category) bool {
	return c.ID == other.ID && c.Name == other.Name
}

// Value 获取分类值（返回自身，值对象本身就是值）
func (c Category) Value() Category {
	return c
}
