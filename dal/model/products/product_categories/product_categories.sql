-- 商品分类关联表 (PostgreSQL 版本)
CREATE TABLE product_categories (
    id SERIAL PRIMARY KEY,
    product_id INTEGER,
    category_id INTEGER,
    UNIQUE (product_id, category_id)
);
