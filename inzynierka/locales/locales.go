package locales

import "embed"

// FS holds the embedded translation JSON files for all supported languages.
//
//go:embed */*.json
var FS embed.FS
