package entity

import (
	"github.com/falconfan123/Go-mall/services/search/internal/domain/valueobject"
)

// QueryIntent represents the parsed search intent
type QueryIntent struct {
	// OriginalQuery is the raw user query string
	OriginalQuery string
	// NormalizedQuery is the normalized form of the query
	NormalizedQuery *valueobject.NormalizedQuery
	// PredictedCategory is the predicted product category
	PredictedCategory string
	// EntityTags contains extracted entity tags (brand, product words, modifiers)
	EntityTags *valueobject.EntityTags
	// DSL is the rewritten DSL fragment for search
	DSL *valueobject.DSLFragment
}

// NewQueryIntent creates a new QueryIntent
func NewQueryIntent(originalQuery string) *QueryIntent {
	return &QueryIntent{
		OriginalQuery:     originalQuery,
		NormalizedQuery:   nil,
		PredictedCategory: "",
		EntityTags:        valueobject.NewEntityTags(),
		DSL:               nil,
	}
}

// ToDSL converts the QueryIntent to DSL fragment
func (q *QueryIntent) ToDSL() *valueobject.DSLFragment {
	if q.DSL == nil {
		return valueobject.NewDSLFragment(q.NormalizedQuery.GetValue(), q.EntityTags)
	}
	return q.DSL
}
