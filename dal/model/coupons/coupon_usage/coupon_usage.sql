-- 优惠券使用记录表 (PostgreSQL 版本)
CREATE TABLE coupon_usage
(
    id              SERIAL PRIMARY KEY,
    order_id        VARCHAR(36)              NOT NULL,
    coupon_id       VARCHAR(36)              NOT NULL,
    user_id         INTEGER             NOT NULL,
    coupon_type     SMALLINT                  NOT NULL,
    origin_value    BIGINT                      NOT NULL,
    discount_amount BIGINT                      NOT NULL,
    applied_at      TIMESTAMP                NOT NULL
);

CREATE INDEX idx_order ON coupon_usage(order_id);
CREATE INDEX idx_user ON coupon_usage(user_id);
