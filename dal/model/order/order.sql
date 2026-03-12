-- 主订单表 (PostgreSQL 版本)
CREATE TABLE orders
(
    order_id        VARCHAR(36)  NOT NULL,
    pre_order_id    VARCHAR(36)  NOT NULL,
    user_id         INTEGER NOT NULL,
    coupon_id       VARCHAR(36)  NOT NULL,

    -- 支付信息
    payment_method  SMALLINT,
    transaction_id  VARCHAR(64),
    paid_at         BIGINT,

    -- 金额信息
    original_amount BIGINT       NOT NULL,
    discount_amount BIGINT       NOT NULL DEFAULT 0,
    payable_amount  BIGINT       NOT NULL,
    paid_amount     BIGINT       DEFAULT NULL,

    -- 状态管理
    order_status    SMALLINT     NOT NULL,
    payment_status  SMALLINT     NOT NULL,

    reason          VARCHAR(255),
    expire_time     BIGINT       NOT NULL,
    created_at      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (order_id),
    UNIQUE (pre_order_id)
);

CREATE INDEX idx_user_status ON orders(user_id, order_status);
CREATE INDEX idx_payment_time ON orders(paid_at);

-- 订单商品快照表 (PostgreSQL 版本)
CREATE TABLE order_items
(
    order_id     VARCHAR(36)  NOT NULL,
    product_id   INTEGER NOT NULL,
    quantity     INTEGER NOT NULL,
    price        BIGINT       NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_desc VARCHAR(255) NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (order_id)
);

CREATE INDEX idx_product ON order_items(product_id);

-- 用户下单地址快照表 (PostgreSQL 版本)
CREATE TABLE order_addresses
(
    order_id         VARCHAR(36)  NOT NULL,
    address_id       BIGSERIAL,
    recipient_name   VARCHAR(100) NOT NULL,
    phone_number     VARCHAR(50)  DEFAULT NULL,
    province         VARCHAR(100) DEFAULT NULL,
    city             VARCHAR(100) NOT NULL,
    detailed_address VARCHAR(255) NOT NULL,
    created_at       TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (address_id),
    UNIQUE (order_id)
);

CREATE INDEX idx_recipient_name ON order_addresses(recipient_name);
