package processor

import (
	"context"

	"github.com/falconfan123/Go-mall/services/search/internal/domain/pipeline"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/valueobject"
)

// Rewriter constructs the DSL fragment based on parsed intent
type Rewriter struct{}

// NewRewriter creates a new Rewriter
func NewRewriter() *Rewriter {
	return &Rewriter{}
}

// Process creates the DSL fragment from the pipeline context
func (r *Rewriter) Process(ctx context.Context, pCtx *pipeline.PipelineContext) error {
	if pCtx == nil {
		return nil
	}

	// Build query string from normalized query
	queryString := ""
	if pCtx.Normalized != nil {
		queryString = pCtx.Normalized.GetValue()
	}

	// Create DSL fragment with entity tags
	dsl := valueobject.NewDSLFragment(queryString, pCtx.Entities)

	// Adjust boost based on category
	if pCtx.Category != "" {
		dsl.Boost = r.calculateBoost(pCtx.Entities, pCtx.Category)
	}

	pCtx.DSL = dsl

	return nil
}

// Name returns the processor name
func (r *Rewriter) Name() string {
	return "Rewriter"
}

// calculateBoost calculates the overall boost based on entities and category
func (r *Rewriter) calculateBoost(entities *valueobject.EntityTags, category string) float64 {
	boost := 1.0

	// Boost for brand matches
	if len(entities.Brands) > 0 {
		boost *= 1.5
	}

	// Boost for category match
	if category != "" {
		boost *= 1.2
	}

	// Slight boost for modifiers
	if len(entities.Modifiers) > 0 {
		boost *= 1.1
	}

	return boost
}
