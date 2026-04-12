-- 优惠券主表 (PostgreSQL 版本)
CREATE TABLE coupons
(
    id              VARCHAR(36)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    type            SMALLINT      NOT NULL,
    value           BIGINT       NOT NULL,
    min_amount      BIGINT       DEFAULT 0,
    start_time      TIMESTAMP    NOT NULL,
    end_time        TIMESTAMP    NOT NULL,
    status          SMALLINT     NOT NULL DEFAULT 1,
    total_count     INTEGER      NOT NULL,
    remaining_count INTEGER      NOT NULL,
    created_at      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE INDEX idx_time ON coupons(start_time, end_time);
CREATE INDEX idx_status ON coupons(status);
