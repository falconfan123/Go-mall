-- 购物车表 (PostgreSQL 版本)
CREATE TABLE carts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER DEFAULT NULL,
    product_id INTEGER DEFAULT NULL,
    quantity INTEGER DEFAULT NULL,
    checked SMALLINT DEFAULT NULL
);

CREATE INDEX idx_carts_user_id ON carts(user_id);
CREATE INDEX idx_carts_product_id ON carts(product_id);
