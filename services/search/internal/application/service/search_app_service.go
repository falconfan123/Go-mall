package service

import (
	"context"

	"github.com/falconfan123/Go-mall/services/search/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/service"
	"github.com/falconfan123/Go-mall/services/search/internal/infrastructure/persistence"
)

// SearchAppService is the application service for search
type SearchAppService struct {
	queryParser *service.QueryParser
	esClient    *persistence.ElasticsearchClient
}

// NewSearchAppService creates a new SearchAppService
func NewSearchAppService(qp *service.QueryParser, esClient *persistence.ElasticsearchClient) *SearchAppService {
	return &SearchAppService{
		queryParser: qp,
		esClient:    esClient,
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
func (s *SearchAppService) Search(ctx context.Context, req *dto.QueryRequest) (*dto.SearchResponse, error) {
	// First parse the query to get intent
	intent, err := s.queryParser.Parse(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	// Build the DSL from intent
	dsl := intent.ToDSL()

	// Execute search against ES
	return s.esClient.Search(ctx, dsl, req)
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
