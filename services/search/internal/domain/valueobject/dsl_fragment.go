package valueobject

// FieldWeight represents a field with its boost weight
type FieldWeight struct {
	Field  string  `json:"field"`
	Weight float64 `json:"weight"`
}

// DSLFragment represents the rewritten DSL for search
type DSLFragment struct {
	// QueryString is the main query string
	QueryString string `json:"query_string"`
	// Fields with their boost weights
	Fields []FieldWeight `json:"fields"`
	// Boost overall boost factor
	Boost float64 `json:"boost"`
}

// NewDSLFragment creates a new DSLFragment
func NewDSLFragment(query string, tags *EntityTags) *DSLFragment {
	fields := make([]FieldWeight, 0)

	// Base fields with default weights
	fields = append(fields, FieldWeight{Field: "name", Weight: 1.0})
	fields = append(fields, FieldWeight{Field: "description", Weight: 0.5})

	// Add brand field if brands are found
	if tags != nil && len(tags.Brands) > 0 {
		fields = append(fields, FieldWeight{Field: "brand", Weight: 2.0})
	}

	// Add category field if product words are found
	if tags != nil && len(tags.ProductWords) > 0 {
		fields = append(fields, FieldWeight{Field: "category", Weight: 1.5})
	}

	return &DSLFragment{
		QueryString: query,
		Fields:      fields,
		Boost:       1.0,
	}
}

// ToMapQuery converts the DSLFragment to Elasticsearch map query format
func (d *DSLFragment) ToMapQuery() map[string]interface{} {
	if d == nil {
		return nil
	}

	// Build fields array for multi_match
	fieldsArray := make([]string, 0, len(d.Fields))
	for _, f := range d.Fields {
		fieldsArray = append(fieldsArray, f.Field+"^"+formatWeight(f.Weight))
	}

	query := map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  d.QueryString,
			"fields": fieldsArray,
			"boost":  d.Boost,
		},
	}

	return query
}

func formatWeight(weight float64) string {
	if weight == float64(int(weight)) {
		return string(rune('0' + int(weight)))
	}
	return ""
}
