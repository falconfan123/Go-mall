package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// SettlementException 异常单人工待办（跨层聚合去重，partial unique index）。
type SettlementException struct {
	Id           int64          `db:"id"`
	OrderId      string         `db:"order_id"`
	PaymentId    sql.NullString `db:"payment_id"`
	Sources      sql.NullString `db:"sources"` // jsonb 数组（["relay","scan1",...]）
	Reason       string         `db:"reason"`
	Detail       sql.NullString `db:"detail"`
	Status       string         `db:"status"` // open|resolved
	FirstSeenAt  sql.NullTime   `db:"first_seen_at"`
	LastSeenAt   sql.NullTime   `db:"last_seen_at"`
	SeenCount    int64          `db:"seen_count"`
	ResolvedAt   sql.NullTime   `db:"resolved_at"`
	ResolvedNote sql.NullString `db:"resolved_note"`
}

var _ SettlementExceptionModel = (*customSettlementExceptionModel)(nil)

type (
	// SettlementExceptionModel 异常单模型。
	SettlementExceptionModel interface {
		// UpsertOpen 跨层聚合：同 order_id 的 open 待办 upsert（sources 去重追加、
		// seen_count+1、last_seen_at 刷新），resolved 后同单新异常开新行。
		// 返回 (seenCount, isNew)：isNew=false 时调用方不重复触发告警（design 决策 2）。
		UpsertOpen(ctx context.Context, orderId, paymentId, source, reason string, detail map[string]any) (int64, bool, error)
		// FindOpenByOrderId 查询同单 open 待办（可查语义）。
		FindOpenByOrderId(ctx context.Context, orderId string) (*SettlementException, error)
		// MarkResolved 标记处置结果（可关语义）。
		MarkResolved(ctx context.Context, orderId, note string) error
	}

	defaultSettlementExceptionModel struct {
		conn  sqlx.SqlConn
		table string
	}

	customSettlementExceptionModel struct {
		defaultSettlementExceptionModel
	}
)

func NewSettlementExceptionModel(conn sqlx.SqlConn) SettlementExceptionModel {
	return &customSettlementExceptionModel{
		defaultSettlementExceptionModel{conn: conn, table: "settlement_exception"},
	}
}

// UpsertOpen 聚合去重写入：partial unique index (order_id WHERE status='open') 冲突时
// 去重追加来源、seen_count+1、刷新 last_seen_at/reason/detail。
func (m *defaultSettlementExceptionModel) UpsertOpen(ctx context.Context, orderId, paymentId, source, reason string, detail map[string]any) (int64, bool, error) {
	detailJSON := "null"
	if detail != nil {
		b, err := json.Marshal(detail)
		if err != nil {
			return 0, false, err
		}
		detailJSON = string(b)
	}
	// sources 去重追加：旧数组与 [source] 合并后 DISTINCT
	query := fmt.Sprintf(`INSERT INTO %s (order_id, payment_id, sources, reason, detail)
		VALUES ($1, NULLIF($2, ''), to_jsonb(ARRAY[$3::text]), $4, $5::jsonb)
		ON CONFLICT (order_id) WHERE status = 'open' DO UPDATE SET
		  sources = (SELECT jsonb_agg(DISTINCT s) FROM jsonb_array_elements(
		      %s.sources || to_jsonb(ARRAY[$3::text])) AS s(s)),
		  seen_count = %s.seen_count + 1,
		  last_seen_at = now(),
		  reason = EXCLUDED.reason,
		  detail = COALESCE(EXCLUDED.detail, %s.detail),
		  payment_id = COALESCE(NULLIF($2, ''), %s.payment_id)
		RETURNING seen_count, (xmax = 0) AS is_new`,
		m.table, m.table, m.table, m.table, m.table)
	var row struct {
		SeenCount int64 `db:"seen_count"`
		IsNew     bool  `db:"is_new"`
	}
	if err := m.conn.QueryRowCtx(ctx, &row, query, orderId, paymentId, source, reason, detailJSON); err != nil {
		return 0, false, err
	}
	return row.SeenCount, row.IsNew, nil
}

// FindOpenByOrderId 可查语义：返回同单 open 待办（无则 ErrNotFound）。
func (m *defaultSettlementExceptionModel) FindOpenByOrderId(ctx context.Context, orderId string) (*SettlementException, error) {
	query := fmt.Sprintf(`SELECT id, order_id, payment_id, sources, reason, detail, status,
		first_seen_at, last_seen_at, seen_count, resolved_at, resolved_note
		FROM %s WHERE order_id = $1 AND status = 'open' LIMIT 1`, m.table)
	var out SettlementException
	err := m.conn.QueryRowCtx(ctx, &out, query, orderId)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// MarkResolved 可关语义：标记处置结果；resolved 后同单新异常开新行。
func (m *defaultSettlementExceptionModel) MarkResolved(ctx context.Context, orderId, note string) error {
	query := fmt.Sprintf(`UPDATE %s SET status = 'resolved', resolved_at = now(), resolved_note = $2
		WHERE order_id = $1 AND status = 'open'`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, orderId, note)
	return err
}
