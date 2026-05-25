package stats

import (
	"context"
	"errors"
	"time"
)

type Service struct {
	store   Store
	metrics *Metrics
	now     func() time.Time
}

func NewService(store Store, metrics *Metrics) *Service {
	return &Service{
		store:   store,
		metrics: metrics,
		now:     time.Now,
	}
}

func (s *Service) EnsureSchema(ctx context.Context) error {
	if s.store == nil {
		return nil
	}
	return s.store.EnsureSchema(ctx)
}

func (s *Service) RecordConversationCreated(ctx context.Context, meta RequestMetadata) (bool, error) {
	if s.store == nil {
		return false, errors.New("stats store is nil")
	}

	created, err := s.store.RecordConversationCreated(ctx, meta, s.now().UTC())
	if err != nil {
		return false, err
	}
	if created && s.metrics != nil {
		s.metrics.IncConversation(meta.Labels(StatusSuccess, false))
	}
	return created, nil
}

func (s *Service) RecordTurnStarted(ctx context.Context, meta RequestMetadata, isRAG bool) error {
	return s.recordEvent(ctx, Event{
		ConversationID:  meta.ConversationID,
		TurnID:          meta.TurnID,
		UserID:          meta.UserID,
		AppID:           meta.AppID,
		KnowledgeBaseID: meta.KnowledgeBaseID,
		Channel:         meta.Channel,
		EventType:       EventTurnStarted,
		IsRAG:           isRAG,
		RAGStrategy:     strategyForEvent(meta, isRAG),
		Status:          StatusStarted,
		TraceID:         meta.TraceID,
		CreatedAt:       s.now().UTC(),
	}, func() {
		if s.metrics != nil {
			s.metrics.IncTurn(meta.Labels(StatusStarted, isRAG))
		}
	})
}

func (s *Service) RecordRAGTriggered(ctx context.Context, meta RequestMetadata) error {
	return s.recordEvent(ctx, Event{
		ConversationID:  meta.ConversationID,
		TurnID:          meta.TurnID,
		UserID:          meta.UserID,
		AppID:           meta.AppID,
		KnowledgeBaseID: meta.KnowledgeBaseID,
		Channel:         meta.Channel,
		EventType:       EventRAGTriggered,
		IsRAG:           true,
		RAGStrategy:     strategyForEvent(meta, true),
		Status:          StatusStarted,
		TraceID:         meta.TraceID,
		CreatedAt:       s.now().UTC(),
	}, func() {
		if s.metrics != nil {
			s.metrics.IncRAGTurn(meta.Labels(StatusStarted, true))
		}
	})
}

func (s *Service) RecordRetrievalCompleted(ctx context.Context, meta RequestMetadata, docCount int, duration time.Duration, status string, errorCode string) error {
	return s.recordEvent(ctx, Event{
		ConversationID:    meta.ConversationID,
		TurnID:            meta.TurnID,
		UserID:            meta.UserID,
		AppID:             meta.AppID,
		KnowledgeBaseID:   meta.KnowledgeBaseID,
		Channel:           meta.Channel,
		EventType:         EventRAGRetrievalDone,
		IsRAG:             true,
		RetrievedDocCount: docCount,
		RAGStrategy:       strategyForEvent(meta, true),
		Status:            status,
		ErrorCode:         errorCode,
		LatencyMS:         duration.Milliseconds(),
		TraceID:           meta.TraceID,
		CreatedAt:         s.now().UTC(),
	}, func() {
		if s.metrics == nil {
			return
		}
		labels := meta.Labels(status, true)
		s.metrics.ObserveRetrievalDuration(labels, duration)
		switch status {
		case StatusSuccess:
			s.metrics.IncRetrievalSuccess(labels)
		case StatusEmpty:
			s.metrics.IncRetrievalEmpty(labels)
		}
	})
}

func (s *Service) RecordAnswerCompleted(ctx context.Context, meta RequestMetadata, duration time.Duration, totalDuration time.Duration, status string, errorCode string) error {
	return s.recordEvent(ctx, Event{
		ConversationID:  meta.ConversationID,
		TurnID:          meta.TurnID,
		UserID:          meta.UserID,
		AppID:           meta.AppID,
		KnowledgeBaseID: meta.KnowledgeBaseID,
		Channel:         meta.Channel,
		EventType:       EventRAGAnswerDone,
		IsRAG:           true,
		RAGStrategy:     strategyForEvent(meta, true),
		Status:          status,
		ErrorCode:       errorCode,
		LatencyMS:       duration.Milliseconds(),
		TraceID:         meta.TraceID,
		CreatedAt:       s.now().UTC(),
	}, func() {
		if s.metrics == nil {
			return
		}
		labels := meta.Labels(status, true)
		s.metrics.ObserveGenerationDuration(labels, duration)
		s.metrics.ObserveTurnDuration(labels, totalDuration)
		if status == StatusSuccess {
			s.metrics.IncAnswerSuccess(labels)
			return
		}
		s.metrics.IncAnswerFailure(labels)
	})
}

func (s *Service) GetOverview(ctx context.Context, filter EventFilter) (Overview, error) {
	return s.store.GetOverview(ctx, filter)
}

func (s *Service) GetTrend(ctx context.Context, filter EventFilter, granularity string) ([]TrendPoint, error) {
	return s.store.GetTrend(ctx, filter, granularity)
}

func (s *Service) GetBreakdown(ctx context.Context, filter EventFilter, dimension string) ([]BreakdownItem, error) {
	return s.store.GetBreakdown(ctx, filter, dimension)
}

func (s *Service) ListEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	return s.store.ListEvents(ctx, filter)
}

func (s *Service) recordEvent(ctx context.Context, event Event, onSuccess func()) error {
	if s.store == nil {
		return errors.New("stats store is nil")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = s.now().UTC()
	}
	if err := s.store.RecordEvent(ctx, event); err != nil {
		return err
	}
	if onSuccess != nil {
		onSuccess()
	}
	return nil
}

func strategyForEvent(meta RequestMetadata, isRAG bool) string {
	if !isRAG {
		return DefaultNonRAGStrategy
	}
	if meta.RAGStrategy == "" {
		return DefaultRAGStrategy
	}
	return meta.RAGStrategy
}
