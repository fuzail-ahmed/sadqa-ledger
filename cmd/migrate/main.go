// Command migrate applies pending database migrations against the
// configured database path. The server binary applies migrations
// automatically on startup (see cmd/server); this exists so `make migrate`
// runs the exact same code path without starting the HTTP server.
package main

import (
	"log/slog"
	"os"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/config"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/db"
	"github.com/fuzail-ahmed/sadqa-ledger/migrations"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	conn, err := db.Open(cfg.DatabasePath)
	if err != nil {
		logger.Error("open database", "path", cfg.DatabasePath, "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := db.Migrate(conn, migrations.FS); err != nil {
		logger.Error("apply migrations", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations applied", "path", cfg.DatabasePath)
}
