module github.com/fuzail-ahmed/sadqa-ledger

go 1.25.0

require (
	github.com/a-h/templ v0.3.1020
	github.com/go-chi/chi/v5 v5.2.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	modernc.org/libc v1.72.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

// modernc.org/sqlite: pure-Go SQLite driver (no cgo), pinned ahead of Phase 1
// (Database & Migrations) even though Phase 0 doesn't import it yet. See
// docs/TRD.md §5 for the reasoning.
require modernc.org/sqlite v1.50.1
