package valueobject

// EntityTags represents extracted entity tags from the query
type EntityTags struct {
	// Brands contains identified brand names
	Brands []string
	// ProductWords contains identified product type words
	ProductWords []string
	// Modifiers contains identified modifier words (color, size, etc.)
	Modifiers []string
}

// NewEntityTags creates a new EntityTags
func NewEntityTags() *EntityTags {
	return &EntityTags{
		Brands:       []string{},
		ProductWords: []string{},
		Modifiers:    []string{},
	}
}

// AddBrand adds a brand to the tags
func (e *EntityTags) AddBrand(brand string) {
	if brand == "" {
		return
	}
	for _, b := range e.Brands {
		if b == brand {
			return
		}
	}
	e.Brands = append(e.Brands, brand)
}

// AddProductWord adds a product word to the tags
func (e *EntityTags) AddProductWord(word string) {
	if word == "" {
		return
	}
	for _, w := range e.ProductWords {
		if w == word {
			return
		}
	}
	e.ProductWords = append(e.ProductWords, word)
}

// AddModifier adds a modifier to the tags
func (e *EntityTags) AddModifier(modifier string) {
	if modifier == "" {
		return
	}
	for _, m := range e.Modifiers {
		if m == modifier {
			return
		}
	}
	e.Modifiers = append(e.Modifiers, modifier)
}

// IsEmpty returns true if no tags are set
func (e *EntityTags) IsEmpty() bool {
	return len(e.Brands) == 0 && len(e.ProductWords) == 0 && len(e.Modifiers) == 0
}
