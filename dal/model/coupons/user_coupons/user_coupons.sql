-- 用户优惠券关联表 (PostgreSQL 版本)
CREATE TABLE user_coupons
(
    id        SERIAL PRIMARY KEY,
    user_id   INTEGER NOT NULL,
    coupon_id VARCHAR(36)  NOT NULL,
    status    SMALLINT NOT NULL DEFAULT 0,
    order_id  VARCHAR(36)  DEFAULT NULL,
    used_at   TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, coupon_id),
    UNIQUE (order_id)
);

CREATE INDEX idx_user_status ON user_coupons(user_id, status);
CREATE INDEX idx_order ON user_coupons(order_id);
