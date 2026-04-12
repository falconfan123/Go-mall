package processor

import (
	"context"

	"github.com/falconfan123/Go-mall/services/search/internal/domain/pipeline"
)

// CategoryPredictor predicts product categories from the query
type CategoryPredictor struct{}

// NewCategoryPredictor creates a new CategoryPredictor
func NewCategoryPredictor() *CategoryPredictor {
	return &CategoryPredictor{}
}

// Process predicts the product category based on query tokens
func (p *CategoryPredictor) Process(ctx context.Context, pCtx *pipeline.PipelineContext) error {
	if pCtx == nil || len(pCtx.Tokens) == 0 {
		return nil
	}

	// Simple category prediction based on keyword matching
	category := p.predictCategory(pCtx.Tokens)
	pCtx.Category = category

	return nil
}

// Name returns the processor name
func (p *CategoryPredictor) Name() string {
	return "CategoryPredictor"
}

// predictCategory predicts the category based on tokens
func (p *CategoryPredictor) predictCategory(tokens []string) string {
	// Category keyword mapping
	categoryKeywords := map[string][]string{
		"electronics": {"phone", "laptop", "tablet", "tv", "camera", "watch", "headphone", "earphone"},
		"appliances":  {"refrigerator", "washing", "machine", "air", "conditioner", "microwave", "oven"},
		"computers":   {"laptop", "computer", "desktop", "monitor", "keyboard", "mouse"},
		"accessories": {"case", "charger", "cable", "cover", " protector"},
	}

	// Count matches for each category
	categoryScores := make(map[string]int)
	for _, token := range tokens {
		for category, keywords := range categoryKeywords {
			for _, keyword := range keywords {
				if token == keyword {
					categoryScores[category]++
				}
			}
		}
	}

	// Find the category with the highest score
	maxScore := 0
	predictedCategory := ""
	for category, score := range categoryScores {
		if score > maxScore {
			maxScore = score
			predictedCategory = category
		}
	}

	return predictedCategory
}
