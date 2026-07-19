// Package server wires up the chi router: middleware, static asset
// serving, and page routes. Phase 2 adds auth (login/logout/first-run
// setup) in front of admin routes (docs/IMPLEMENTATION_PLAN.md).
package server

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/config"
	"github.com/fuzail-ahmed/sadqa-ledger/web/static"
)

// New builds the application's HTTP handler.
func New(conn *sql.DB, cfg config.Config) http.Handler {
	h := &authHandlers{conn: conn, cfg: cfg}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Embedded static assets (compiled Tailwind CSS + vendored Basecoat
	// bundle) — see web/static/embed.go.
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))

	// Public auth routes.
	r.Get("/setup", h.handleSetupPage)
	r.Post("/setup", h.handleSetupSubmit)
	r.Get("/login", h.handleLoginPage)
	r.Post("/login", h.handleLoginSubmit)

	// Admin routes, behind session auth (docs/SCHEMA.md §6).
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(conn, cfg.SecureCookies))
		r.Get("/", h.handleHome)
		r.Post("/logout", h.handleLogout)
		r.Post("/lang", h.handleLangSubmit)
		r.Get("/admins/new", h.handleAdminNewPage)
		r.Post("/admins/new", h.handleAdminNewSubmit)

		// Route scaffolding for the remaining admin screens
		// (docs/APP_FLOW.md §0) — placeholder content only, real logic lands
		// in the phase that owns each screen (docs/IMPLEMENTATION_PLAN.md).
		r.Get("/members", h.handleMembersPage)
		r.Get("/members/new", h.handleMemberNewPage)
		r.Post("/members/new", h.handleMemberNewSubmit)
		r.Get("/members/{id}/edit", h.handleMemberEditPage)
		r.Post("/members/{id}/edit", h.handleMemberEditSubmit)
		r.Post("/members/{id}/toggle", h.handleMemberToggle)
		r.Get("/contributions", h.handlePlaceholder("more", "nav.contributions"))
		r.Get("/contributions/new", h.handlePlaceholder("add", "nav.add_contribution"))
		r.Get("/expenses", h.handlePlaceholder("expenses", "nav.expenses"))
		r.Get("/expenses/new", h.handlePlaceholder("expenses", "expenses.new_title"))
		r.Get("/summary", h.handlePlaceholder("more", "nav.summary"))
		r.Get("/settings", h.handlePlaceholder("more", "nav.settings"))
		r.Get("/export", h.handlePlaceholder("more", "nav.export"))
	})

	return r
}
