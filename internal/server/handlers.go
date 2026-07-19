package server

import (
	"net/http"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

// handleHome renders the Phase 0 placeholder page, now behind auth
// (Phase 2). Real Dashboard content lands in Phase 5.
func (h *authHandlers) handleHome(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	csrfToken := auth.CSRFToken(w, r, h.cfg.SecureCookies)
	pages.Home(admin.DisplayName, csrfToken).Render(r.Context(), w)
}
