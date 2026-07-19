package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/dashboard"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/settings"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

// handlePublicPage renders the public read-only transparency page (docs/APP_FLOW.md §9).
func (h *authHandlers) handlePublicPage(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.renderPublic404(w, r)
		return
	}

	gs, err := settings.Get(h.conn)
	if err != nil {
		h.renderPublic404(w, r)
		return
	}

	// Verify token match (docs/APP_FLOW.md §9)
	if gs.PublicToken != token {
		h.renderPublic404(w, r)
		return
	}

	// Fetch dashboard data
	currentMonth := time.Now().Format("2006-01")
	data, err := dashboard.GetDashboardData(h.conn, currentMonth)
	if err != nil {
		h.renderPublic404(w, r)
		return
	}

	// Set robots noindex headers (docs/TRD.md §9)
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	dir := "ltr"
	if gs.DefaultPublicLanguage == "ar" {
		dir = "rtl"
	}
	pages.PublicPage(gs.DefaultPublicLanguage, dir, gs, data).Render(r.Context(), w)
}

func (h *authHandlers) renderPublic404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Generic 404 response specified in docs/APP_FLOW.md §9
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8"/>
	<title>Not Found</title>
	<meta name="robots" content="noindex, nofollow"/>
	<style>
		body { font-family: system-ui, sans-serif; text-align: center; padding: 50px; background: #FAF9F6; color: #1F2422; }
		h1 { font-size: 24px; margin-bottom: 10px; }
		p { color: #6B7280; }
	</style>
</head>
<body>
	<h1>This page isn't available.</h1>
	<p>Please confirm you have the correct link.</p>
</body>
</html>`))
}
