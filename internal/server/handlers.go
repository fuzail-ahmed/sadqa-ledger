package server

import (
	"net/http"

	"github.com/fuzail-ahmed/sadqa-ledger/i18n"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/components"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

// shellData gathers the values every admin screen's shared header/nav needs
// (docs/UI_UX_BRIEF.md §5/§7, §8 bilingual/RTL) from the current request's
// authenticated admin, so each handler doesn't repeat the lookup. titleKey is
// an i18n key resolved into the page's <title>.
func (h *authHandlers) shellData(w http.ResponseWriter, r *http.Request, active, titleKey string) components.AdminShellData {
	admin := auth.CurrentAdmin(r)
	dir := "ltr"
	if admin.LanguagePref == "ar" {
		dir = "rtl"
	}
	return components.AdminShellData{
		Lang:      admin.LanguagePref,
		Dir:       dir,
		CSRFToken: auth.CSRFToken(w, r, h.cfg.SecureCookies),
		Active:    active,
		Title:     i18n.T(admin.LanguagePref, titleKey),
	}
}

// handleHome renders the Dashboard route, behind auth (Phase 2). Real
// Dashboard content (stat cards, checklist, activity feed) lands in Phase 5.
func (h *authHandlers) handleHome(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	pages.Home(h.shellData(w, r, "dashboard", "nav.dashboard"), admin.DisplayName).Render(r.Context(), w)
}

// handleLangSubmit switches the current admin's UI language
// (docs/APP_FLOW.md §10: "available from every screen ... applies
// immediately"), persisted on the admin row (docs/SCHEMA.md §3.1
// language_pref) so it follows them across devices/sessions.
func (h *authHandlers) handleLangSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	lang := r.FormValue("lang")
	if lang != "en" && lang != "ar" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	admin := auth.CurrentAdmin(r)
	if err := auth.SetLanguagePref(h.conn, admin.ID, lang); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	redirectTo := r.Referer()
	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// handlePlaceholder renders a not-yet-implemented admin screen scaffolded in
// Phase 3 (docs/IMPLEMENTATION_PLAN.md); its real content lands in the phase
// that owns it (docs/APP_FLOW.md §0 route map).
func (h *authHandlers) handlePlaceholder(active, headingKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pages.Placeholder(h.shellData(w, r, active, headingKey), headingKey).Render(r.Context(), w)
	}
}
