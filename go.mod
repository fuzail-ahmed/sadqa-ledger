module github.com/fuzail-ahmed/sadqa-ledger

go 1.26.4

// modernc.org/sqlite: pure-Go SQLite driver (no cgo). See docs/TRD.md §5 for
// the reasoning. Version verified at https://pkg.go.dev/modernc.org/sqlite
// (latest listed release, 2026-07-19).
require modernc.org/sqlite v1.50.1
