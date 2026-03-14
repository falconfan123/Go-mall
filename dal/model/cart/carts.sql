-- 购物车表 (PostgreSQL 版本)
CREATE TABLE IF NOT EXISTS carts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER DEFAULT NULL,
    product_id INTEGER DEFAULT NULL,
    product_name VARCHAR(255) DEFAULT NULL,
    product_image VARCHAR(512) DEFAULT NULL,
    product_price DECIMAL(10, 2) DEFAULT NULL,
    quantity INTEGER DEFAULT NULL,
    checked SMALLINT DEFAULT NULL
);

-- 如果表已存在，添加新列
-- ALTER TABLE carts ADD COLUMN IF NOT EXISTS product_name VARCHAR(255) DEFAULT NULL;
-- ALTER TABLE carts ADD COLUMN IF NOT EXISTS product_image VARCHAR(512) DEFAULT NULL;
-- ALTER TABLE carts ADD COLUMN IF NOT EXISTS product_price DECIMAL(10, 2) DEFAULT NULL;

CREATE INDEX idx_carts_user_id ON carts(user_id);
CREATE INDEX idx_carts_product_id ON carts(product_id);
