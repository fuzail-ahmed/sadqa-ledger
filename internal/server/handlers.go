package server

import (
	"net/http"

	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

// handleHome renders the Phase 0 placeholder page.
func handleHome(w http.ResponseWriter, r *http.Request) {
	pages.Home().Render(r.Context(), w)
}
