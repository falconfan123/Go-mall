package service

import (
	"context"

	"github.com/falconfan123/Go-mall/services/search/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/pipeline"
	"github.com/falconfan123/Go-mall/services/search/internal/infrastructure/processor"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"golang.org/x/sync/singleflight"
)

// QueryParser is the domain service for parsing search queries
type QueryParser struct {
	executor    *pipeline.PipelineExecutor
	redisClient redis.RedisConf
	// Cache for query results
	cacheGroup singleflight.Group
}

// NewQueryParser creates a new QueryParser
func NewQueryParser(redisConf redis.RedisConf) *QueryParser {
	// Create pipeline executor with all processors
	exec := pipeline.NewPipelineExecutor(
		// Preprocessor (OrderPreprocessor)
		processor.NewPreprocessor(),
		// EntityExtractor (OrderEntityExtractor)
		processor.NewEntityExtractor(),
		// CategoryPredictor (OrderCategoryPredictor)
		processor.NewCategoryPredictor(),
		// Rewriter (OrderRewriter)
		processor.NewRewriter(),
	)

	return &QueryParser{
		executor:    exec,
		redisClient: redisConf,
	}
}

// Parse parses a search query and returns a QueryIntent
func (p *QueryParser) Parse(ctx context.Context, query string) (*entity.QueryIntent, error) {
	if query == "" {
		return nil, nil
	}

	// Create pipeline context
	pCtx := pipeline.NewPipelineContext(ctx, query)

	// Execute the pipeline
	if err := p.executor.Execute(ctx, pCtx); err != nil {
		return nil, err
	}

	// Build the QueryIntent from pipeline context
	intent := entity.NewQueryIntent(query)
	intent.NormalizedQuery = pCtx.Normalized
	intent.PredictedCategory = pCtx.Category
	intent.EntityTags = pCtx.Entities
	intent.DSL = pCtx.DSL

	return intent, nil
}

// ParseWithCache parses a query with caching support
func (p *QueryParser) ParseWithCache(ctx context.Context, query string) (*entity.QueryIntent, error) {
	// TODO: Implement caching with Redis
	// For now, just use Parse
	return p.Parse(ctx, query)
}
