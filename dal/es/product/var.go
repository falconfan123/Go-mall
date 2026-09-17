package product

import _ "embed"

//go:embed mapping.json
var EsMapping string

//go:embed template.json
var EsTemplate string
