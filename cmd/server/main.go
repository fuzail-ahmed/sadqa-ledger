// Command server is Sadqa Ledger's HTTP entrypoint. It starts a single
// static binary serving the chi-routed, templ-rendered app with embedded
// static assets (see internal/server and web/static).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/config"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/db"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/server"
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

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           server.New(conn, cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Periodic sweep of expired sessions (docs/SCHEMA.md §5); lookups also
	// lazily delete an expired row the moment it's used, so this just
	// catches sessions nobody has tried to use since they expired.
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := auth.DeleteExpiredSessions(conn); err != nil {
					logger.Error("cleanup expired sessions", "error", err)
				} else if n > 0 {
					logger.Info("cleaned up expired sessions", "count", n)
				}
			}
		}
	}()

	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
