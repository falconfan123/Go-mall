-- 支付表 (PostgreSQL 版本)
CREATE TABLE payments
(
    payment_id      VARCHAR(36)  NOT NULL,
    pre_order_id    VARCHAR(64)  NOT NULL,
    order_id        VARCHAR(36)  DEFAULT NULL,
    user_id         INTEGER NOT NULL,
    original_amount BIGINT       NOT NULL,
    paid_amount     BIGINT       DEFAULT NULL,

    -- 支付信息
    payment_method  VARCHAR(20)  NOT NULL,
    transaction_id  VARCHAR(64)  DEFAULT NULL,
    pay_url         TEXT         NOT NULL,
    expire_time     BIGINT       NOT NULL,

    status          SMALLINT      NOT NULL,

    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at         BIGINT       DEFAULT NULL,

    PRIMARY KEY (payment_id)
);

CREATE INDEX idx_pre_order ON payments(pre_order_id);
CREATE INDEX idx_order ON payments(order_id);
CREATE INDEX idx_status_method ON payments(status, payment_method);
CREATE INDEX idx_create_time ON payments(created_at);
