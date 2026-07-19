// Package server wires up the chi router: middleware, static asset
// serving, and page routes. Phase 0 only serves one placeholder page; later
// phases add auth, the dashboard, contributions, expenses, etc.
// (docs/IMPLEMENTATION_PLAN.md).
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/fuzail-ahmed/sadqa-ledger/web/static"
)

// New builds the application's HTTP handler.
func New() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Embedded static assets (compiled Tailwind CSS + vendored Basecoat
	// bundle) — see web/static/embed.go.
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))

	r.Get("/", handleHome)

	return r
}
