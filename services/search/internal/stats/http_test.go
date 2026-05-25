package stats

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestHTTPHandlerServesStatsAndMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewMetricsForRegistry(registry)
	store := newMemoryStore()
	service := NewService(store, metrics)
	service.now = func() time.Time {
		return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	}

	meta := RequestMetadata{
		ConversationID:  "conv-http",
		TurnID:          "turn-http",
		AppID:           "search-admin",
		KnowledgeBaseID: "products",
		Channel:         "grpc",
		RAGStrategy:     "elasticsearch",
	}
	if _, err := service.RecordConversationCreated(context.Background(), meta); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordTurnStarted(context.Background(), meta, false); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordRAGTriggered(context.Background(), meta); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordRetrievalCompleted(context.Background(), meta, 0, 20*time.Millisecond, StatusEmpty, ""); err != nil {
		t.Fatal(err)
	}
	if err := service.RecordAnswerCompleted(context.Background(), meta, 5*time.Millisecond, 30*time.Millisecond, StatusFailure, "gen_failed"); err != nil {
		t.Fatal(err)
	}

	handler := NewHTTPHandler("/metrics", service, registry)

	req := httptest.NewRequest(http.MethodGet, "/stats/rag/overview", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	var overview Overview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatal(err)
	}
	if overview.TotalConversations != 1 || overview.TotalTurns != 1 || overview.RAGTurns != 1 {
		t.Fatalf("unexpected overview: %+v", overview)
	}

	req = httptest.NewRequest(http.MethodGet, "/stats/rag/events?app_id=search-admin&limit=10", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected events status: %d", rec.Code)
	}
	var events []Event
	if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}

	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected metrics status: %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "rag_turn_rag_total") || !strings.Contains(body, "rag_answer_failure_total") {
		t.Fatalf("metrics output missing rag counters: %s", body)
	}
}
