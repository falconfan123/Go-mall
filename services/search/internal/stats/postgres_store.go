package stats

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) EnsureSchema(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("postgres db is nil")
	}

	const ddl = `
CREATE TABLE IF NOT EXISTS rag_stat_events (
  id BIGSERIAL PRIMARY KEY,
  conversation_id VARCHAR(128) NOT NULL,
  turn_id VARCHAR(128) NOT NULL,
  user_id VARCHAR(128) NOT NULL DEFAULT '',
  app_id VARCHAR(128) NOT NULL,
  knowledge_base_id VARCHAR(128) NOT NULL,
  channel VARCHAR(64) NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  is_rag BOOLEAN NOT NULL DEFAULT FALSE,
  retrieved_doc_count INTEGER NOT NULL DEFAULT 0,
  rag_strategy VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT '',
  error_code VARCHAR(128) NOT NULL DEFAULT '',
  latency_ms BIGINT NOT NULL DEFAULT 0,
  trace_id VARCHAR(128) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_rag_stat_conversation_created
  ON rag_stat_events(conversation_id)
  WHERE event_type = 'conversation_created';
CREATE INDEX IF NOT EXISTS idx_rag_stat_created_at ON rag_stat_events(created_at);
CREATE INDEX IF NOT EXISTS idx_rag_stat_app_created_at ON rag_stat_events(app_id, created_at);
CREATE INDEX IF NOT EXISTS idx_rag_stat_kb_created_at ON rag_stat_events(knowledge_base_id, created_at);
CREATE INDEX IF NOT EXISTS idx_rag_stat_channel_created_at ON rag_stat_events(channel, created_at);
CREATE INDEX IF NOT EXISTS idx_rag_stat_event_type_created_at ON rag_stat_events(event_type, created_at);
CREATE INDEX IF NOT EXISTS idx_rag_stat_turn_id ON rag_stat_events(turn_id);
`
	_, err := s.db.ExecContext(ctx, ddl)
	return err
}

func (s *PostgresStore) RecordConversationCreated(ctx context.Context, meta RequestMetadata, createdAt time.Time) (bool, error) {
	const query = `
INSERT INTO rag_stat_events (
  conversation_id, turn_id, user_id, app_id, knowledge_base_id, channel,
  event_type, is_rag, retrieved_doc_count, rag_strategy, status, error_code, latency_ms, trace_id, created_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
ON CONFLICT DO NOTHING`
	result, err := s.db.ExecContext(
		ctx,
		query,
		meta.ConversationID,
		meta.TurnID,
		meta.UserID,
		meta.AppID,
		meta.KnowledgeBaseID,
		meta.Channel,
		EventConversationCreated,
		false,
		0,
		DefaultNonRAGStrategy,
		StatusSuccess,
		"",
		0,
		meta.TraceID,
		createdAt,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (s *PostgresStore) RecordEvent(ctx context.Context, event Event) error {
	const query = `
INSERT INTO rag_stat_events (
  conversation_id, turn_id, user_id, app_id, knowledge_base_id, channel,
  event_type, is_rag, retrieved_doc_count, rag_strategy, status, error_code, latency_ms, trace_id, created_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
	_, err := s.db.ExecContext(
		ctx,
		query,
		event.ConversationID,
		event.TurnID,
		event.UserID,
		event.AppID,
		event.KnowledgeBaseID,
		event.Channel,
		event.EventType,
		event.IsRAG,
		event.RetrievedDocCount,
		event.RAGStrategy,
		event.Status,
		event.ErrorCode,
		event.LatencyMS,
		event.TraceID,
		event.CreatedAt,
	)
	return err
}

func (s *PostgresStore) GetOverview(ctx context.Context, filter EventFilter) (Overview, error) {
	where, args := buildWhereClause(filter)
	query := `
SELECT
  COUNT(*) FILTER (WHERE event_type = 'conversation_created') AS total_conversations,
  COUNT(*) FILTER (WHERE event_type = 'turn_started') AS total_turns,
  COUNT(*) FILTER (WHERE event_type = 'rag_triggered') AS rag_turns,
  COUNT(*) FILTER (WHERE event_type = 'rag_retrieval_completed' AND status = 'success') AS rag_retrieval_success,
  COUNT(*) FILTER (WHERE event_type = 'rag_retrieval_completed' AND status = 'empty') AS rag_retrieval_empty,
  COUNT(*) FILTER (WHERE event_type = 'rag_answer_completed' AND status = 'success') AS rag_answer_success,
  COUNT(*) FILTER (WHERE event_type = 'rag_answer_completed' AND status = 'failure') AS rag_answer_failure
FROM rag_stat_events` + where

	var overview Overview
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&overview.TotalConversations,
		&overview.TotalTurns,
		&overview.RAGTurns,
		&overview.RAGRetrievalSuccess,
		&overview.RAGRetrievalEmpty,
		&overview.RAGAnswerSuccess,
		&overview.RAGAnswerFailure,
	); err != nil {
		return Overview{}, err
	}
	applyRatios(&overview)
	return overview, nil
}

func (s *PostgresStore) GetTrend(ctx context.Context, filter EventFilter, granularity string) ([]TrendPoint, error) {
	granularity = normalizeGranularity(granularity)
	where, args := buildWhereClause(filter)
	query := fmt.Sprintf(`
SELECT
  date_trunc('%s', created_at) AS bucket,
  COUNT(*) FILTER (WHERE event_type = 'conversation_created') AS total_conversations,
  COUNT(*) FILTER (WHERE event_type = 'turn_started') AS total_turns,
  COUNT(*) FILTER (WHERE event_type = 'rag_triggered') AS rag_turns,
  COUNT(*) FILTER (WHERE event_type = 'rag_retrieval_completed' AND status = 'empty') AS rag_retrieval_empty,
  COUNT(*) FILTER (WHERE event_type = 'rag_answer_completed' AND status = 'success') AS rag_answer_success
FROM rag_stat_events%s
GROUP BY 1
ORDER BY 1`, granularity, where)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TrendPoint
	for rows.Next() {
		var point TrendPoint
		if err := rows.Scan(
			&point.Bucket,
			&point.TotalConversations,
			&point.TotalTurns,
			&point.RAGTurns,
			&point.RAGRetrievalEmpty,
			&point.RAGAnswerSuccess,
		); err != nil {
			return nil, err
		}
		applyTrendRatios(&point)
		points = append(points, point)
	}
	return points, rows.Err()
}

func (s *PostgresStore) GetBreakdown(ctx context.Context, filter EventFilter, dimension string) ([]BreakdownItem, error) {
	dimension = normalizeDimension(dimension)
	where, args := buildWhereClause(filter)
	query := fmt.Sprintf(`
SELECT
  COALESCE(NULLIF(%[1]s, ''), 'unknown') AS bucket_value,
  COUNT(*) FILTER (WHERE event_type = 'conversation_created') AS total_conversations,
  COUNT(*) FILTER (WHERE event_type = 'turn_started') AS total_turns,
  COUNT(*) FILTER (WHERE event_type = 'rag_triggered') AS rag_turns,
  COUNT(*) FILTER (WHERE event_type = 'rag_retrieval_completed' AND status = 'empty') AS rag_retrieval_empty,
  COUNT(*) FILTER (WHERE event_type = 'rag_answer_completed' AND status = 'success') AS rag_answer_success
FROM rag_stat_events%[2]s
GROUP BY 1
ORDER BY total_turns DESC, bucket_value ASC`, dimension, where)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []BreakdownItem
	for rows.Next() {
		var item BreakdownItem
		item.Dimension = dimension
		if err := rows.Scan(
			&item.Value,
			&item.TotalConversations,
			&item.TotalTurns,
			&item.RAGTurns,
			&item.RAGRetrievalEmpty,
			&item.RAGAnswerSuccess,
		); err != nil {
			return nil, err
		}
		applyBreakdownRatios(&item)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) ListEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	where, args := buildWhereClause(filter)
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)
	query := `
SELECT
  id, conversation_id, turn_id, user_id, app_id, knowledge_base_id, channel,
  event_type, is_rag, retrieved_doc_count, rag_strategy, status, error_code, latency_ms, trace_id, created_at
FROM rag_stat_events` + where + fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.ID,
			&event.ConversationID,
			&event.TurnID,
			&event.UserID,
			&event.AppID,
			&event.KnowledgeBaseID,
			&event.Channel,
			&event.EventType,
			&event.IsRAG,
			&event.RetrievedDocCount,
			&event.RAGStrategy,
			&event.Status,
			&event.ErrorCode,
			&event.LatencyMS,
			&event.TraceID,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func buildWhereClause(filter EventFilter) (string, []any) {
	conditions := make([]string, 0, 8)
	args := make([]any, 0, 8)

	appendCond := func(expr string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(expr, len(args)))
	}

	if filter.StartAt != nil {
		appendCond("created_at >= $%d", *filter.StartAt)
	}
	if filter.EndAt != nil {
		appendCond("created_at <= $%d", *filter.EndAt)
	}
	if filter.AppID != "" {
		appendCond("app_id = $%d", filter.AppID)
	}
	if filter.KnowledgeBaseID != "" {
		appendCond("knowledge_base_id = $%d", filter.KnowledgeBaseID)
	}
	if filter.Channel != "" {
		appendCond("channel = $%d", filter.Channel)
	}
	if filter.EventType != "" {
		appendCond("event_type = $%d", filter.EventType)
	}
	if filter.Status != "" {
		appendCond("status = $%d", filter.Status)
	}
	if filter.IsRAG != nil {
		appendCond("is_rag = $%d", *filter.IsRAG)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func normalizeGranularity(granularity string) string {
	switch granularity {
	case "minute", "hour", "day":
		return granularity
	default:
		return "hour"
	}
}

func normalizeDimension(dimension string) string {
	switch dimension {
	case "app_id", "knowledge_base_id", "channel":
		return dimension
	default:
		return "app_id"
	}
}

func applyRatios(overview *Overview) {
	overview.RAGTurnRatio = ratio(overview.RAGTurns, overview.TotalTurns)
	overview.RAGEmptyRatio = ratio(overview.RAGRetrievalEmpty, overview.RAGTurns)
	overview.RAGAnswerSuccessRatio = ratio(overview.RAGAnswerSuccess, overview.RAGTurns)
}

func applyTrendRatios(point *TrendPoint) {
	point.RAGTurnRatio = ratio(point.RAGTurns, point.TotalTurns)
	point.RAGEmptyRatio = ratio(point.RAGRetrievalEmpty, point.RAGTurns)
	point.RAGAnswerSuccessRate = ratio(point.RAGAnswerSuccess, point.RAGTurns)
}

func applyBreakdownRatios(item *BreakdownItem) {
	item.RAGTurnRatio = ratio(item.RAGTurns, item.TotalTurns)
	item.RAGEmptyRatio = ratio(item.RAGRetrievalEmpty, item.RAGTurns)
	item.RAGAnswerSuccessRate = ratio(item.RAGAnswerSuccess, item.RAGTurns)
}

func ratio(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
