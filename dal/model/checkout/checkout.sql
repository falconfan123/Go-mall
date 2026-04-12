-- 预订单表 (PostgreSQL 版本)
CREATE TABLE checkouts
(
    pre_order_id    VARCHAR(64)  NOT NULL,
    user_id         BIGINT NOT NULL,
    address_id      BIGINT NOT NULL,
    coupon_id       VARCHAR(255) DEFAULT NULL,
    original_amount BIGINT       NOT NULL,
    final_amount    BIGINT       NOT NULL,
    status          SMALLINT   NOT NULL DEFAULT 0,
    expire_time     BIGINT       NOT NULL,
    created_at      TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (pre_order_id)
);

CREATE INDEX idx_user_status ON checkouts(user_id, status);
CREATE INDEX idx_expire ON checkouts(expire_time);

-- 预订单商品明细表 (PostgreSQL 版本)
CREATE TABLE checkout_items
(
    id           BIGSERIAL,
    pre_order_id VARCHAR(64)  NOT NULL,
    product_id   INTEGER NOT NULL,
    quantity     INTEGER NOT NULL,
    price        BIGINT       NOT NULL,
    snapshot     JSONB         NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE INDEX idx_product ON checkout_items(product_id);
CREATE INDEX idx_preorder ON checkout_items(pre_order_id);
