package processor

import (
	"context"
	"strings"

	"github.com/falconfan123/Go-mall/services/search/internal/domain/pipeline"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/valueobject"
)

// Preprocessor cleans, escapes, and tokenizes the input query
type Preprocessor struct{}

// NewPreprocessor creates a new Preprocessor
func NewPreprocessor() *Preprocessor {
	return &Preprocessor{}
}

// Process normalizes and tokenizes the query
func (p *Preprocessor) Process(ctx context.Context, pCtx *pipeline.PipelineContext) error {
	if pCtx == nil || pCtx.OriginalQuery == "" {
		return nil
	}

	// Step 1: Clean the query - remove special characters
	cleaned := p.cleanQuery(pCtx.OriginalQuery)

	// Step 2: Create normalized query
	normalized := strings.TrimSpace(strings.ToLower(cleaned))
	pCtx.Normalized = valueobject.NewNormalizedQuery(normalized)

	// Step 3: Tokenize
	tokens := pCtx.Normalized.Tokens()
	pCtx.Tokens = tokens

	return nil
}

// Name returns the processor name
func (p *Preprocessor) Name() string {
	return "Preprocessor"
}

// cleanQuery removes special characters that shouldn't be in search queries
func (p *Preprocessor) cleanQuery(query string) string {
	// Remove HTML tags
	query = strings.ReplaceAll(query, "<", "")
	query = strings.ReplaceAll(query, ">", "")

	// Replace multiple spaces with single space
	query = strings.Join(strings.Fields(query), " ")

	// Trim
	return strings.TrimSpace(query)
}
