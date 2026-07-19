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

	// Local uploaded receipt photos serving route (docs/SCHEMA.md §7, §Assumptions).
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Public auth routes.
	r.Get("/setup", h.handleSetupPage)
	r.Post("/setup", h.handleSetupSubmit)
	r.Get("/login", h.handleLoginPage)
	r.Post("/login", h.handleLoginSubmit)
	r.Get("/p/{token}", h.handlePublicPage)

	// PWA routes served from root (docs/TRD.md §11)
	r.Get("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		data, err := static.FS.ReadFile("manifest.json")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})
	r.Get("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		data, err := static.FS.ReadFile("sw.js")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write(data)
	})

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
		r.Get("/contributions/new", h.handleContributionNewPage)
		r.Post("/contributions/new", h.handleContributionNewSubmit)
		r.Get("/contributions/search-members", h.handleContributionSearchMembers)
		r.Get("/contributions/select-member", h.handleContributionSelectMember)
		r.Get("/contributions/clear-member", h.handleContributionClearMember)
		r.Get("/contributions/check-duplicate", h.handleContributionCheckDuplicate)
		r.Get("/expenses", h.handleExpensesPage)
		r.Get("/expenses/new", h.handleExpenseNewPage)
		r.Post("/expenses/new", h.handleExpenseNewSubmit)
		r.Post("/expenses/{id}/delete", h.handleExpenseDelete)
		r.Get("/summary", h.handleSummaryPage)
		r.Get("/summary/generate", h.handleSummaryGenerate)
		r.Get("/settings", h.handleSettingsPage)
		r.Post("/settings/privacy", h.handleSettingsPrivacySubmit)
		r.Post("/settings/regenerate-token", h.handleSettingsTokenRegenerateSubmit)
		r.Post("/settings/info", h.handleSettingsInfoSubmit)
		r.Post("/settings/language", h.handleSettingsLanguageSubmit)
		r.Get("/export", h.handleExportPage)
		r.Get("/export/database", h.handleExportDatabase)
		r.Get("/export/contributions", h.handleExportContributions)
		r.Get("/export/expenses", h.handleExportExpenses)
	})

	return r
}
