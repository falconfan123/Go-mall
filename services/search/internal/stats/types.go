package stats

import (
	"context"
	"time"
)

const (
	EventConversationCreated = "conversation_created"
	EventTurnStarted         = "turn_started"
	EventRAGTriggered        = "rag_triggered"
	EventRAGRetrievalDone    = "rag_retrieval_completed"
	EventRAGAnswerDone       = "rag_answer_completed"
	StatusStarted            = "started"
	StatusSuccess            = "success"
	StatusEmpty              = "empty"
	StatusFailure            = "failure"
	DefaultChannel           = "grpc"
	DefaultAppID             = "search"
	DefaultRAGStrategy       = "elasticsearch"
	DefaultNonRAGStrategy    = "none"
)

type RequestMetadata struct {
	ConversationID  string
	TurnID          string
	UserID          string
	AppID           string
	KnowledgeBaseID string
	Channel         string
	RAGStrategy     string
	TraceID         string
}

type MetricLabels struct {
	AppID           string
	KnowledgeBaseID string
	Channel         string
	RAGStrategy     string
	Status          string
}

type Event struct {
	ID                int64     `json:"id"`
	ConversationID    string    `json:"conversation_id"`
	TurnID            string    `json:"turn_id"`
	UserID            string    `json:"user_id"`
	AppID             string    `json:"app_id"`
	KnowledgeBaseID   string    `json:"knowledge_base_id"`
	Channel           string    `json:"channel"`
	EventType         string    `json:"event_type"`
	IsRAG             bool      `json:"is_rag"`
	RetrievedDocCount int       `json:"retrieved_doc_count"`
	RAGStrategy       string    `json:"rag_strategy"`
	Status            string    `json:"status"`
	ErrorCode         string    `json:"error_code"`
	LatencyMS         int64     `json:"latency_ms"`
	TraceID           string    `json:"trace_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type EventFilter struct {
	StartAt         *time.Time
	EndAt           *time.Time
	AppID           string
	KnowledgeBaseID string
	Channel         string
	EventType       string
	Status          string
	IsRAG           *bool
	Limit           int
	Offset          int
}

type Overview struct {
	TotalConversations    int64   `json:"total_conversations"`
	TotalTurns            int64   `json:"total_turns"`
	RAGTurns              int64   `json:"rag_turns"`
	RAGRetrievalSuccess   int64   `json:"rag_retrieval_success"`
	RAGRetrievalEmpty     int64   `json:"rag_retrieval_empty"`
	RAGAnswerSuccess      int64   `json:"rag_answer_success"`
	RAGAnswerFailure      int64   `json:"rag_answer_failure"`
	RAGTurnRatio          float64 `json:"rag_turn_ratio"`
	RAGEmptyRatio         float64 `json:"rag_empty_ratio"`
	RAGAnswerSuccessRatio float64 `json:"rag_answer_success_ratio"`
}

type TrendPoint struct {
	Bucket               time.Time `json:"bucket"`
	TotalConversations   int64     `json:"total_conversations"`
	TotalTurns           int64     `json:"total_turns"`
	RAGTurns             int64     `json:"rag_turns"`
	RAGRetrievalEmpty    int64     `json:"rag_retrieval_empty"`
	RAGAnswerSuccess     int64     `json:"rag_answer_success"`
	RAGTurnRatio         float64   `json:"rag_turn_ratio"`
	RAGEmptyRatio        float64   `json:"rag_empty_ratio"`
	RAGAnswerSuccessRate float64   `json:"rag_answer_success_ratio"`
}

type BreakdownItem struct {
	Dimension            string  `json:"dimension"`
	Value                string  `json:"value"`
	TotalConversations   int64   `json:"total_conversations"`
	TotalTurns           int64   `json:"total_turns"`
	RAGTurns             int64   `json:"rag_turns"`
	RAGRetrievalEmpty    int64   `json:"rag_retrieval_empty"`
	RAGAnswerSuccess     int64   `json:"rag_answer_success"`
	RAGTurnRatio         float64 `json:"rag_turn_ratio"`
	RAGEmptyRatio        float64 `json:"rag_empty_ratio"`
	RAGAnswerSuccessRate float64 `json:"rag_answer_success_ratio"`
}

type Store interface {
	EnsureSchema(ctx context.Context) error
	RecordConversationCreated(ctx context.Context, meta RequestMetadata, createdAt time.Time) (bool, error)
	RecordEvent(ctx context.Context, event Event) error
	GetOverview(ctx context.Context, filter EventFilter) (Overview, error)
	GetTrend(ctx context.Context, filter EventFilter, granularity string) ([]TrendPoint, error)
	GetBreakdown(ctx context.Context, filter EventFilter, dimension string) ([]BreakdownItem, error)
	ListEvents(ctx context.Context, filter EventFilter) ([]Event, error)
}
