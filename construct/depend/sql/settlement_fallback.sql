-- ============================================================================
-- settlement_fallback.sql — 结算兜底三层基建（change: settlement-fallback-layers）
--
-- 对 mall 库执行（与 payments 同库，outbox 事务才成立）：
--   payment_outbox     结算待办（webhook 翻转事务内写入，relay 接力 EnsureSaga）
--   settlement_exception 异常单人工待办（跨层聚合去重，partial unique index）
--
-- 时间列统一 timestamptz + 参数化比较（防 naive 字符串时区坑）。
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 结算待办（Outbox）：status = pending|done|dead
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payment_outbox (
  id             bigserial PRIMARY KEY,
  order_id       varchar(64)  NOT NULL,
  payment_id     varchar(64)  NOT NULL,
  user_id        bigint       NOT NULL,
  payload        jsonb        NOT NULL,
  status         varchar(16)  NOT NULL DEFAULT 'pending',
  attempts       int          NOT NULL DEFAULT 0,
  next_retry_at  timestamptz,
  done_at        timestamptz,
  last_error     text,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_outbox_order ON payment_outbox (order_id);
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON payment_outbox (next_retry_at) WHERE status = 'pending';

-- ---------------------------------------------------------------------------
-- 异常单人工待办：status = open|resolved
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS settlement_exception (
  id            bigserial PRIMARY KEY,
  order_id      varchar(64) NOT NULL,
  payment_id    varchar(64),
  sources       jsonb       NOT NULL DEFAULT '[]',
  reason        varchar(255) NOT NULL,
  detail        jsonb,
  status        varchar(16) NOT NULL DEFAULT 'open',
  first_seen_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at  timestamptz NOT NULL DEFAULT now(),
  seen_count    int         NOT NULL DEFAULT 1,
  resolved_at   timestamptz,
  resolved_note varchar(255)
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_exception_open_order ON settlement_exception (order_id) WHERE status = 'open';
