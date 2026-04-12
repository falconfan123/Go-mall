package processor

import (
	"context"
	"strings"

	"github.com/falconfan123/Go-mall/services/search/internal/domain/pipeline"
	"github.com/falconfan123/Go-mall/services/search/internal/domain/valueobject"
)

// brandKeywords are common brand names for matching
var brandKeywords = map[string]bool{
	"apple": true, "samsung": true, "huawei": true, "xiaomi": true,
	"sony": true, "lg": true, "bosch": true, "siemens": true,
	"dell": true, "hp": true, "lenovo": true, "asus": true,
}

// modifierKeywords are common modifier words
var modifierKeywords = map[string]bool{
	"new": true, "hot": true, "sale": true, "discount": true,
	"cheap": true, "expensive": true, "premium": true, "budget": true,
	"mini": true, "pro": true, "plus": true, "max": true,
	"lite": true, "ultra": true,
}

// productKeywords are common product type words
var productKeywords = map[string]bool{
	"phone": true, "mobile": true, "laptop": true, "computer": true,
	"tablet": true, "watch": true, "headphone": true, "earphone": true,
	"camera": true, "tv": true, "television": true, "refrigerator": true,
	"washing": true, "machine": true, "air": true, "conditioner": true,
}

// EntityExtractor extracts brand, product words, and modifiers from tokens
type EntityExtractor struct{}

// NewEntityExtractor creates a new EntityExtractor
func NewEntityExtractor() *EntityExtractor {
	return &EntityExtractor{}
}

// Process extracts entity tags from the tokenized query
func (e *EntityExtractor) Process(ctx context.Context, pCtx *pipeline.PipelineContext) error {
	if pCtx == nil || len(pCtx.Tokens) == 0 {
		return nil
	}

	// Ensure entities are initialized
	if pCtx.Entities == nil {
		pCtx.Entities = valueobject.NewEntityTags()
	}

	// Process each token
	for _, token := range pCtx.Tokens {
		// Check for brand
		if brandKeywords[token] {
			pCtx.Entities.AddBrand(token)
			continue
		}

		// Check for modifier
		if modifierKeywords[token] {
			pCtx.Entities.AddModifier(token)
			continue
		}

		// Check for product word
		if productKeywords[token] || e.isProductCategory(token) {
			pCtx.Entities.AddProductWord(token)
		}
	}

	// Also try to extract brand from compound queries like "iphone 13"
	// This is a simple implementation - can be enhanced with more sophisticated matching
	e.extractCompoundBrands(pCtx.Tokens, pCtx.Entities)

	return nil
}

// Name returns the processor name
func (e *EntityExtractor) Name() string {
	return "EntityExtractor"
}

// isProductCategory checks if a token is a product category
func (e *EntityExtractor) isProductCategory(token string) bool {
	// Simple category detection based on common patterns
	productPatterns := []string{"phone", "laptop", "tv", "watch", "camera", "tablet"}
	for _, pattern := range productPatterns {
		if strings.Contains(token, pattern) {
			return true
		}
	}
	return false
}

// extractCompoundBrands handles compound brand names like "iphone 13"
func (e *EntityExtractor) extractCompoundBrands(tokens []string, entities *valueobject.EntityTags) {
	if len(tokens) < 2 {
		return
	}

	// Check if first token is a known brand followed by a model number
	if len(tokens) >= 2 {
		first := tokens[0]
		second := tokens[1]

		// Known brand + model pattern
		if (first == "iphone" || first == "ipad" || first == "macbook") &&
			(len(second) <= 4 && len(second) >= 2) {
			// Treat as brand
			entities.AddBrand(first)
		}
	}
}
