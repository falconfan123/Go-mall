package stats

import (
	"context"
	"strings"

	"github.com/falconfan123/Go-mall/common/utils/metadatactx"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

func ExtractRequestMetadata(ctx context.Context, defaultKnowledgeBaseID string) RequestMetadata {
	meta := RequestMetadata{
		ConversationID:  firstMetadata(ctx, "conversation_id", "x-conversation-id"),
		TurnID:          firstMetadata(ctx, "turn_id", "x-turn-id"),
		UserID:          firstMetadata(ctx, "user_id", "x-user-id"),
		AppID:           firstMetadata(ctx, "app_id", "x-app-id"),
		KnowledgeBaseID: firstMetadata(ctx, "knowledge_base_id", "x-knowledge-base-id"),
		Channel:         firstMetadata(ctx, "channel", "x-channel"),
		RAGStrategy:     firstMetadata(ctx, "rag_strategy", "x-rag-strategy"),
		TraceID:         firstMetadata(ctx, "trace_id", "x-trace-id"),
	}

	if meta.ConversationID == "" {
		meta.ConversationID = uuid.NewString()
	}
	if meta.TurnID == "" {
		meta.TurnID = uuid.NewString()
	}
	if meta.AppID == "" {
		meta.AppID = DefaultAppID
	}
	if meta.KnowledgeBaseID == "" {
		meta.KnowledgeBaseID = defaultKnowledgeBaseID
	}
	if meta.Channel == "" {
		meta.Channel = DefaultChannel
	}
	if meta.RAGStrategy == "" {
		meta.RAGStrategy = DefaultRAGStrategy
	}
	if meta.TraceID == "" {
		if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
			meta.TraceID = span.TraceID().String()
		}
	}

	return meta
}

func (m RequestMetadata) Labels(status string, isRAG bool) MetricLabels {
	strategy := m.RAGStrategy
	if !isRAG {
		strategy = DefaultNonRAGStrategy
	}

	return MetricLabels{
		AppID:           emptyToUnknown(m.AppID),
		KnowledgeBaseID: emptyToUnknown(m.KnowledgeBaseID),
		Channel:         emptyToUnknown(m.Channel),
		RAGStrategy:     emptyToUnknown(strategy),
		Status:          emptyToUnknown(status),
	}
}

func firstMetadata(ctx context.Context, keys ...string) string {
	for _, key := range keys {
		if value, ok := metadatactx.ExtractFromMetadataCtx(ctx, key); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func emptyToUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}
