package scaffold

import "embed"

// TemplatesFS holds the embedded service scaffolding templates. The "all:"
// prefix is required so that dot-prefixed paths such as .github are included.
//
//go:embed all:templates
var TemplatesFS embed.FS
