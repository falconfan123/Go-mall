package pipeline

import (
	"context"

	"github.com/falconfan123/Go-mall/services/search/internal/domain/valueobject"
)

// PipelineContext holds the state during pipeline execution
type PipelineContext struct {
	// OriginalQuery is the raw user query string
	OriginalQuery string
	// Normalized is the normalized form of the query
	Normalized *valueobject.NormalizedQuery
	// Tokens contains the tokenized query terms
	Tokens []string
	// Entities contains extracted entity tags
	Entities *valueobject.EntityTags
	// Category is the predicted product category
	Category string
	// DSL is the rewritten DSL fragment
	DSL *valueobject.DSLFragment
	// Err holds any error that occurred during processing
	Err error

	// Internal context for cancellation
	ctx context.Context
}

// NewPipelineContext creates a new PipelineContext
func NewPipelineContext(ctx context.Context, query string) *PipelineContext {
	return &PipelineContext{
		OriginalQuery: query,
		Normalized:    nil,
		Tokens:        []string{},
		Entities:      valueobject.NewEntityTags(),
		Category:      "",
		DSL:           nil,
		Err:           nil,
		ctx:           ctx,
	}
}

// Context returns the internal context
func (p *PipelineContext) Context() context.Context {
	return p.ctx
}

// HasError returns true if an error occurred
func (p *PipelineContext) HasError() bool {
	return p.Err != nil
}

// SetError sets an error and marks the context as failed
func (p *PipelineContext) SetError(err error) {
	if p.Err == nil && err != nil {
		p.Err = err
	}
}

// IsCancelled checks if the context was cancelled
func (p *PipelineContext) IsCancelled() bool {
	select {
	case <-p.ctx.Done():
		return true
	default:
		return false
	}
}
