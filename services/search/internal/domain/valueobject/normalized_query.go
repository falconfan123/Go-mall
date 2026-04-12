package valueobject

import "strings"

// NormalizedQuery represents the normalized form of a search query
type NormalizedQuery struct {
	value string
}

// NewNormalizedQuery creates a new NormalizedQuery
func NewNormalizedQuery(value string) *NormalizedQuery {
	// Basic normalization: lowercase and trim whitespace
	normalized := strings.TrimSpace(strings.ToLower(value))
	return &NormalizedQuery{value: normalized}
}

// GetValue returns the normalized query value
func (n *NormalizedQuery) GetValue() string {
	if n == nil {
		return ""
	}
	return n.value
}

// Tokens returns the tokens from normalized query
func (n *NormalizedQuery) Tokens() []string {
	if n == nil || n.value == "" {
		return []string{}
	}
	return strings.Fields(n.value)
}
