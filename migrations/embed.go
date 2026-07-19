// Package migrations embeds the project's numbered .sql migration files into
// the binary so a self-hoster upgrading to a new release only needs to
// restart (docs/SCHEMA.md §8) — no separate migration step or file to copy.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
