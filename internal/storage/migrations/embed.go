package migrations

import "embed"

// Files contains versioned SQL migration scripts.
//
//go:embed *.sql
var Files embed.FS
