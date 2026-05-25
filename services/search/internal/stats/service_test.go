package stats

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	promtest "github.com/prometheus/client_golang/prometheus/testutil"
)

type memoryStore struct {
	events        []Event
	conversations map[string]struct{}
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		conversations: make(map[string]struct{}),
	}
}

func (m *memoryStore) EnsureSchema(context.Context) error { return nil }

func (m *memoryStore) RecordConversationCreated(_ context.Context, meta RequestMetadata, createdAt time.Time) (bool, error) {
	if _, ok := m.conversations[meta.ConversationID]; ok {
		return false, nil
	}
	m.conversations[meta.ConversationID] = struct{}{}
	m.events = append(m.events, Event{
		ConversationID:  meta.ConversationID,
		TurnID:          meta.TurnID,
		UserID:          meta.UserID,
		AppID:           meta.AppID,
		KnowledgeBaseID: meta.KnowledgeBaseID,
		Channel:         meta.Channel,
		EventType:       EventConversationCreated,
		Status:          StatusSuccess,
		CreatedAt:       createdAt,
	})
	return true, nil
}

func (m *memoryStore) RecordEvent(_ context.Context, event Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *memoryStore) GetOverview(_ context.Context, filter EventFilter) (Overview, error) {
	var overview Overview
	for _, event := range filterEvents(m.events, filter) {
		switch {
		case event.EventType == EventConversationCreated:
			overview.TotalConversations++
		case event.EventType == EventTurnStarted:
			overview.TotalTurns++
		case event.EventType == EventRAGTriggered:
			overview.RAGTurns++
		case event.EventType == EventRAGRetrievalDone && event.Status == StatusSuccess:
			overview.RAGRetrievalSuccess++
		case event.EventType == EventRAGRetrievalDone && event.Status == StatusEmpty:
			overview.RAGRetrievalEmpty++
		case event.EventType == EventRAGAnswerDone && event.Status == StatusSuccess:
			overview.RAGAnswerSuccess++
		case event.EventType == EventRAGAnswerDone && event.Status == StatusFailure:
			overview.RAGAnswerFailure++
		}
	}
	applyRatios(&overview)
	return overview, nil
}

func (m *memoryStore) GetTrend(_ context.Context, filter EventFilter, granularity string) ([]TrendPoint, error) {
	buckets := map[time.Time]*TrendPoint{}
	for _, event := range filterEvents(m.events, filter) {
		bucket := event.CreatedAt.Truncate(granularityDuration(normalizeGranularity(granularity)))
		point, ok := buckets[bucket]
		if !ok {
			point = &TrendPoint{Bucket: bucket}
			buckets[bucket] = point
		}
		switch {
		case event.EventType == EventConversationCreated:
			point.TotalConversations++
		case event.EventType == EventTurnStarted:
			point.TotalTurns++
		case event.EventType == EventRAGTriggered:
			point.RAGTurns++
		case event.EventType == EventRAGRetrievalDone && event.Status == StatusEmpty:
			point.RAGRetrievalEmpty++
		case event.EventType == EventRAGAnswerDone && event.Status == StatusSuccess:
			point.RAGAnswerSuccess++
		}
	}
	points := make([]TrendPoint, 0, len(buckets))
	for _, point := range buckets {
		applyTrendRatios(point)
		points = append(points, *point)
	}
	return points, nil
}

func (m *memoryStore) GetBreakdown(_ context.Context, filter EventFilter, dimension string) ([]BreakdownItem, error) {
	key := normalizeDimension(dimension)
	items := map[string]*BreakdownItem{}
	for _, event := range filterEvents(m.events, filter) {
		value := breakdownValue(event, key)
		item, ok := items[value]
		if !ok {
			item = &BreakdownItem{Dimension: key, Value: value}
			items[value] = item
		}
		switch {
		case event.EventType == EventConversationCreated:
			item.TotalConversations++
		case event.EventType == EventTurnStarted:
			item.TotalTurns++
		case event.EventType == EventRAGTriggered:
			item.RAGTurns++
		case event.EventType == EventRAGRetrievalDone && event.Status == StatusEmpty:
			item.RAGRetrievalEmpty++
		case event.EventType == EventRAGAnswerDone && event.Status == StatusSuccess:
			item.RAGAnswerSuccess++
		}
	}
	result := make([]BreakdownItem, 0, len(items))
	for _, item := range items {
		applyBreakdownRatios(item)
		result = append(result, *item)
	}
	return result, nil
}

func (m *memoryStore) ListEvents(_ context.Context, filter EventFilter) ([]Event, error) {
	return filterEvents(m.events, filter), nil
}

func TestServiceRecordsMetricsAndOverview(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewMetricsForRegistry(registry)
	store := newMemoryStore()
	service := NewService(store, metrics)
	service.now = func() time.Time {
		return time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	}

	meta := RequestMetadata{
		ConversationID:  "conv-1",
		TurnID:          "turn-1",
		AppID:           "search-app",
		KnowledgeBaseID: "products",
		Channel:         "grpc",
		RAGStrategy:     "elasticsearch",
	}

	if _, err := service.RecordConversationCreated(context.Background(), meta); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RecordConversationCreated(context.Background(), meta); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordTurnStarted(context.Background(), meta, false); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordTurnStarted(context.Background(), RequestMetadata{
		ConversationID:  "conv-1",
		TurnID:          "turn-2",
		AppID:           "search-app",
		KnowledgeBaseID: "products",
		Channel:         "grpc",
	}, false); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordRAGTriggered(context.Background(), meta); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordRetrievalCompleted(context.Background(), meta, 2, 40*time.Millisecond, StatusSuccess, ""); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordAnswerCompleted(context.Background(), meta, 10*time.Millisecond, 60*time.Millisecond, StatusSuccess, ""); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordTurnStarted(context.Background(), RequestMetadata{
		ConversationID:  "conv-2",
		TurnID:          "turn-3",
		AppID:           "search-app",
		KnowledgeBaseID: "products",
		Channel:         "grpc",
		RAGStrategy:     "elasticsearch",
	}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RecordConversationCreated(context.Background(), RequestMetadata{
		ConversationID:  "conv-2",
		TurnID:          "turn-3",
		AppID:           "search-app",
		KnowledgeBaseID: "products",
		Channel:         "grpc",
	}); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordRAGTriggered(context.Background(), RequestMetadata{
		ConversationID:  "conv-2",
		TurnID:          "turn-3",
		AppID:           "search-app",
		KnowledgeBaseID: "products",
		Channel:         "grpc",
		RAGStrategy:     "elasticsearch",
	}); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordRetrievalCompleted(context.Background(), RequestMetadata{
		ConversationID:  "conv-2",
		TurnID:          "turn-3",
		AppID:           "search-app",
		KnowledgeBaseID: "products",
		Channel:         "grpc",
		RAGStrategy:     "elasticsearch",
	}, 0, 30*time.Millisecond, StatusEmpty, ""); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordAnswerCompleted(context.Background(), RequestMetadata{
		ConversationID:  "conv-2",
		TurnID:          "turn-3",
		AppID:           "search-app",
		KnowledgeBaseID: "products",
		Channel:         "grpc",
		RAGStrategy:     "elasticsearch",
	}, 5*time.Millisecond, 35*time.Millisecond, StatusFailure, "llm_failed"); err != nil {
		t.Fatal(err)
	}

	overview, err := service.GetOverview(context.Background(), EventFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if overview.TotalConversations != 2 {
		t.Fatalf("expected 2 conversations, got %d", overview.TotalConversations)
	}
	if overview.TotalTurns != 3 {
		t.Fatalf("expected 3 turns, got %d", overview.TotalTurns)
	}
	if overview.RAGTurns != 2 {
		t.Fatalf("expected 2 rag turns, got %d", overview.RAGTurns)
	}
	if overview.RAGRetrievalSuccess != 1 || overview.RAGRetrievalEmpty != 1 {
		t.Fatalf("unexpected retrieval counts: %+v", overview)
	}
	if overview.RAGAnswerSuccess != 1 || overview.RAGAnswerFailure != 1 {
		t.Fatalf("unexpected answer counts: %+v", overview)
	}

	if got := promtest.ToFloat64(metrics.conversations.WithLabelValues("search-app", "products", "grpc", "none", "success")); got != 2 {
		t.Fatalf("expected 2 conversations metric, got %v", got)
	}
	if got := promtest.ToFloat64(metrics.turns.WithLabelValues("search-app", "products", "grpc", "none", "started")); got != 3 {
		t.Fatalf("expected 3 turn metric, got %v", got)
	}
	if got := promtest.ToFloat64(metrics.ragTurns.WithLabelValues("search-app", "products", "grpc", "elasticsearch", "started")); got != 2 {
		t.Fatalf("expected 2 rag turn metric, got %v", got)
	}
	if got := promtest.ToFloat64(metrics.retrievalSuccess.WithLabelValues("search-app", "products", "grpc", "elasticsearch", "success")); got != 1 {
		t.Fatalf("expected retrieval success metric 1, got %v", got)
	}
	if got := promtest.ToFloat64(metrics.retrievalEmpty.WithLabelValues("search-app", "products", "grpc", "elasticsearch", "empty")); got != 1 {
		t.Fatalf("expected retrieval empty metric 1, got %v", got)
	}
	if got := promtest.ToFloat64(metrics.answerFailure.WithLabelValues("search-app", "products", "grpc", "elasticsearch", "failure")); got != 1 {
		t.Fatalf("expected answer failure metric 1, got %v", got)
	}
}

func filterEvents(events []Event, filter EventFilter) []Event {
	result := make([]Event, 0, len(events))
	for _, event := range events {
		if filter.StartAt != nil && event.CreatedAt.Before(*filter.StartAt) {
			continue
		}
		if filter.EndAt != nil && event.CreatedAt.After(*filter.EndAt) {
			continue
		}
		if filter.AppID != "" && event.AppID != filter.AppID {
			continue
		}
		if filter.KnowledgeBaseID != "" && event.KnowledgeBaseID != filter.KnowledgeBaseID {
			continue
		}
		if filter.Channel != "" && event.Channel != filter.Channel {
			continue
		}
		if filter.EventType != "" && event.EventType != filter.EventType {
			continue
		}
		if filter.Status != "" && event.Status != filter.Status {
			continue
		}
		if filter.IsRAG != nil && event.IsRAG != *filter.IsRAG {
			continue
		}
		result = append(result, event)
	}
	if filter.Offset >= len(result) {
		return []Event{}
	}
	start := filter.Offset
	end := len(result)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}
	return result[start:end]
}

func granularityDuration(granularity string) time.Duration {
	switch granularity {
	case "minute":
		return time.Minute
	case "day":
		return 24 * time.Hour
	default:
		return time.Hour
	}
}

func breakdownValue(event Event, dimension string) string {
	switch dimension {
	case "knowledge_base_id":
		return emptyToUnknown(event.KnowledgeBaseID)
	case "channel":
		return emptyToUnknown(event.Channel)
	default:
		return emptyToUnknown(event.AppID)
	}
}
