package service

import (
	"context"
	"time"

	"github.com/falconfan123/Go-mall/services/search/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/service"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/valueobject"
	"github.com/falconfan123/Go-mall/services/search/internal/infrastructure/persistence"
	"github.com/falconfan123/Go-mall/services/search/internal/stats"
)

// SearchAppService is the application service for search
type SearchAppService struct {
	queryParser *service.QueryParser
	esClient    *persistence.ElasticsearchClient
	stats       *stats.Service
}

// NewSearchAppService creates a new SearchAppService
func NewSearchAppService(qp *service.QueryParser, esClient *persistence.ElasticsearchClient, statsService *stats.Service) *SearchAppService {
	return &SearchAppService{
		queryParser: qp,
		esClient:    esClient,
		stats:       statsService,
	}
}

// ParseQuery parses a query and returns the intent
func (s *SearchAppService) ParseQuery(ctx context.Context, query string) (*dto.QueryIntentDTO, error) {
	intent, err := s.queryParser.Parse(ctx, query)
	if err != nil {
		return nil, err
	}

	if intent == nil {
		return nil, nil
	}

	return &dto.QueryIntentDTO{
		OriginalQuery:     intent.OriginalQuery,
		NormalizedQuery:   intent.NormalizedQuery.GetValue(),
		PredictedCategory: intent.PredictedCategory,
		Brands:            intent.EntityTags.Brands,
		ProductWords:      intent.EntityTags.ProductWords,
		Modifiers:         intent.EntityTags.Modifiers,
	}, nil
}

// Search performs a search with the given request
func (s *SearchAppService) Search(ctx context.Context, req *dto.QueryRequest, meta stats.RequestMetadata) (*dto.SearchResponse, error) {
	totalStart := time.Now()
	if _, err := s.stats.RecordConversationCreated(ctx, meta); err != nil {
		return nil, err
	}
	if err := s.stats.RecordTurnStarted(ctx, meta, false); err != nil {
		return nil, err
	}
	if err := s.stats.RecordRAGTriggered(ctx, meta); err != nil {
		return nil, err
	}

	// First parse the query to get intent
	intent, err := s.queryParser.Parse(ctx, req.Query)
	if err != nil {
		_ = s.stats.RecordAnswerCompleted(ctx, meta, 0, time.Since(totalStart), stats.StatusFailure, "parse_query_failed")
		return nil, err
	}

	// Build the DSL from intent
	var dsl = (*valueobject.DSLFragment)(nil)
	if intent != nil {
		dsl = intent.ToDSL()
	}

	// Execute search against ES
	retrievalStart := time.Now()
	response, err := s.esClient.Search(ctx, dsl, req)
	retrievalDuration := time.Since(retrievalStart)
	if err != nil {
		_ = s.stats.RecordRetrievalCompleted(ctx, meta, 0, retrievalDuration, stats.StatusFailure, "retrieval_failed")
		_ = s.stats.RecordAnswerCompleted(ctx, meta, 0, time.Since(totalStart), stats.StatusFailure, "retrieval_failed")
		return nil, err
	}

	retrievedDocCount := len(response.Results)
	retrievalStatus := stats.StatusSuccess
	if retrievedDocCount == 0 {
		retrievalStatus = stats.StatusEmpty
	}
	if err := s.stats.RecordRetrievalCompleted(ctx, meta, retrievedDocCount, retrievalDuration, retrievalStatus, ""); err != nil {
		return nil, err
	}

	generationDuration := time.Since(retrievalStart) - retrievalDuration
	if generationDuration < 0 {
		generationDuration = 0
	}
	if err := s.stats.RecordAnswerCompleted(ctx, meta, generationDuration, time.Since(totalStart), stats.StatusSuccess, ""); err != nil {
		return nil, err
	}
	return response, nil
}

// SearchByIntent performs a search using the parsed intent
func (s *SearchAppService) SearchByIntent(ctx context.Context, query string, req *dto.QueryRequest) (*dto.SearchResponse, error) {
	intent, err := s.queryParser.Parse(ctx, query)
	if err != nil {
		return nil, err
	}

	// Use category from request if provided
	category := req.Category
	if category == "" && intent.PredictedCategory != "" {
		category = intent.PredictedCategory
	}

	// Execute search
	return s.esClient.SearchWithCategory(ctx, intent.DSL, category, req)
}
