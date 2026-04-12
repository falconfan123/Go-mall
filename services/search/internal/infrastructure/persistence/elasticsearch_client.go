package persistence

import (
	"context"
	"encoding/json"

	"github.com/falconfan123/Go-mall/common/config"
	"github.com/falconfan123/Go-mall/services/search/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/valueobject"
	"github.com/olivere/elastic/v7"
)

// ElasticsearchClient handles Elasticsearch operations
type ElasticsearchClient struct {
	client    *elastic.Client
	indexName string
}

// NewElasticsearchClient creates a new ElasticsearchClient
func NewElasticsearchClient(cfg config.ElasticSearchConfig) *ElasticsearchClient {
	// Create ES client
	client, err := elastic.NewClient(
		elastic.SetURL(cfg.Addr),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(true),
	)
	if err != nil {
		// In production, handle this error properly
		// For now, we'll return a client that will fail gracefully
		return &ElasticsearchClient{
			client:    nil,
			indexName: cfg.IndexName,
		}
	}

	return &ElasticsearchClient{
		client:    client,
		indexName: cfg.IndexName,
	}
}

// Search performs a search using the DSL fragment
func (c *ElasticsearchClient) Search(ctx context.Context, dsl *valueobject.DSLFragment, req *dto.QueryRequest) (*dto.SearchResponse, error) {
	if c.client == nil {
		return &dto.SearchResponse{
			Results:    []dto.SearchResult{},
			Total:      0,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: 0,
		}, nil
	}

	// Build the query - convert map to elastic.Query
	query := c.buildQueryFromDSL(dsl)
	if query == nil {
		query = elastic.NewMatchAllQuery()
	}

	// Execute search
	from := (req.Page - 1) * req.PageSize
	result, err := c.client.Search().
		Index(c.indexName).
		Query(query).
		From(from).
		Size(req.PageSize).
		TrackTotalHits(true).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	// Parse results
	return c.parseResults(result, req)
}

// SearchWithCategory performs a search with category filter
func (c *ElasticsearchClient) SearchWithCategory(ctx context.Context, dsl *valueobject.DSLFragment, category string, req *dto.QueryRequest) (*dto.SearchResponse, error) {
	if c.client == nil {
		return &dto.SearchResponse{
			Results:    []dto.SearchResult{},
			Total:      0,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: 0,
		}, nil
	}

	// Build the base query
	baseQuery := c.buildQueryFromDSL(dsl)
	if baseQuery == nil {
		baseQuery = elastic.NewMatchAllQuery()
	}

	// Add category filter if provided
	var query elastic.Query
	if category != "" {
		boolQuery := elastic.NewBoolQuery().Must(baseQuery)
		boolQuery = boolQuery.Filter(elastic.NewTermQuery("category", category))
		query = boolQuery
	} else {
		query = baseQuery
	}

	// Execute search
	from := (req.Page - 1) * req.PageSize
	result, err := c.client.Search().
		Index(c.indexName).
		Query(query).
		From(from).
		Size(req.PageSize).
		TrackTotalHits(true).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	// Parse results
	return c.parseResults(result, req)
}

// buildQueryFromDSL converts DSLFragment to elastic.Query
func (c *ElasticsearchClient) buildQueryFromDSL(dsl *valueobject.DSLFragment) elastic.Query {
	if dsl == nil || dsl.QueryString == "" {
		return nil
	}

	// Build fields for multi_match
	fields := make([]string, 0, len(dsl.Fields))
	for _, f := range dsl.Fields {
		fields = append(fields, f.Field+"^"+formatWeight(f.Weight))
	}

	// Create multi_match query
	query := elastic.NewMultiMatchQuery(dsl.QueryString, fields...)
	if dsl.Boost != 1.0 {
		query = query.Boost(dsl.Boost)
	}

	return query
}

func formatWeight(weight float64) string {
	if weight == float64(int(weight)) {
		return string(rune('0' + int(weight)))
	}
	return ""
}

// parseResults parses ES search results
func (c *ElasticsearchClient) parseResults(result *elastic.SearchResult, req *dto.QueryRequest) (*dto.SearchResponse, error) {
	results := make([]dto.SearchResult, 0)

	if result.Hits != nil && result.Hits.Hits != nil {
		for _, hit := range result.Hits.Hits {
			var product dto.SearchResult
			if err := json.Unmarshal(hit.Source, &product); err != nil {
				continue
			}
			if hit.Score != nil {
				product.Score = *hit.Score
			}
			results = append(results, product)
		}
	}

	total := int64(0)
	if result.Hits != nil && result.Hits.TotalHits != nil {
		total = result.Hits.TotalHits.Value
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &dto.SearchResponse{
		Results:    results,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}
