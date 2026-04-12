-- 审计日志表 (PostgreSQL 版本)
CREATE TABLE audit
(
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER not null,
    action_type  VARCHAR(64)  not null,
    action_desc  TEXT,
    old_data     JSONB,
    new_data     JSONB,
    service_name VARCHAR(64)  not null,
    target_table VARCHAR(64)  not null,
    target_id    INTEGER not null,
    client_ip    VARCHAR(45)  not null,
    trace_id     VARCHAR(36)  not null,
    span_id      VARCHAR(36)  not null,
    created_at   TIMESTAMP default CURRENT_TIMESTAMP,
    UNIQUE (trace_id)
);

CREATE INDEX idx_user ON audit(user_id);
CREATE INDEX idx_service ON audit(service_name);
CREATE INDEX idx_action ON audit(action_type);
CREATE INDEX idx_target ON audit(target_table, target_id);
CREATE INDEX idx_time ON audit(created_at);
