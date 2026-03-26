package migrations

import "embed"

// Files embeds SQL migration files so the API can initialize the schema at startup.
//
//go:embed *.sql
var Files embed.FS
