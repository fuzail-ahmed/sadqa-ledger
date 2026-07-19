module github.com/fuzail-ahmed/sadqa-ledger

go 1.25.0

require (
	github.com/a-h/templ v0.3.1020
	github.com/go-chi/chi/v5 v5.2.1
)

// modernc.org/sqlite: pure-Go SQLite driver (no cgo), pinned ahead of Phase 1
// (Database & Migrations) even though Phase 0 doesn't import it yet. See
// docs/TRD.md §5 for the reasoning.
require modernc.org/sqlite v1.50.1
