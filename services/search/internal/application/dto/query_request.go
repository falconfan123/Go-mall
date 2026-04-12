package dto

// QueryRequest represents the search query request
type QueryRequest struct {
	// Query is the search query string
	Query string `json:"query"`
	// Category is the optional category filter
	Category string `json:"category,omitempty"`
	// Page is the page number (1-based)
	Page int `json:"page"`
	// PageSize is the number of results per page
	PageSize int `json:"page_size"`
	// SortBy is the sort field
	SortBy string `json:"sort_by,omitempty"`
	// SortOrder is the sort order (asc/desc)
	SortOrder string `json:"sort_order,omitempty"`
}

// SearchResponse represents the search response
type SearchResponse struct {
	// Results is the list of search results
	Results []SearchResult `json:"results"`
	// Total is the total number of results
	Total int64 `json:"total"`
	// Page is the current page
	Page int `json:"page"`
	// PageSize is the page size
	PageSize int `json:"page_size"`
	// TotalPages is the total number of pages
	TotalPages int `json:"total_pages"`
}

// SearchResult represents a single search result
type SearchResult struct {
	// ID is the product ID
	ID int64 `json:"id"`
	// Name is the product name
	Name string `json:"name"`
	// Description is the product description
	Description string `json:"description"`
	// Price is the product price
	Price float64 `json:"price"`
	// ImageURL is the product image URL
	ImageURL string `json:"image_url,omitempty"`
	// Category is the product category
	Category string `json:"category,omitempty"`
	// Brand is the product brand
	Brand string `json:"brand,omitempty"`
	// Score is the search relevance score
	Score float64 `json:"score,omitempty"`
}

// QueryIntentDTO represents the parsed query intent for API response
type QueryIntentDTO struct {
	// OriginalQuery is the original query string
	OriginalQuery string `json:"original_query"`
	// NormalizedQuery is the normalized query string
	NormalizedQuery string `json:"normalized_query"`
	// PredictedCategory is the predicted category
	PredictedCategory string `json:"predicted_category"`
	// Brands contains extracted brands
	Brands []string `json:"brands"`
	// ProductWords contains extracted product words
	ProductWords []string `json:"product_words"`
	// Modifiers contains extracted modifiers
	Modifiers []string `json:"modifiers"`
}
