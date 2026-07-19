package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/settings"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

func getBaseUrl(r *http.Request) string {
	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseUrl = fmt.Sprintf("%s://%s", scheme, r.Host)
	}
	return strings.TrimSuffix(baseUrl, "/")
}

// handleSettingsPage renders the settings screen.
func (h *authHandlers) handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	admins, err := auth.ListAdmins(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	toast := r.URL.Query().Get("toast")
	pages.Settings(
		h.shellData(w, r, "more", "nav.settings"),
		gs,
		admins,
		getBaseUrl(r),
		toast,
	).Render(r.Context(), w)
}

// handleSettingsPrivacySubmit updates the privacy toggle.
func (h *authHandlers) handleSettingsPrivacySubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	gs.ShowNamesPublicly = r.FormValue("show_names_publicly") == "1"

	if err := settings.Update(h.conn, gs); err != nil {
		http.Error(w, "failed to update settings", http.StatusInternalServerError)
		return
	}

	successMsg := "Privacy settings saved."
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		pages.SettingsPrivacy(
			admin.LanguagePref,
			auth.CSRFToken(w, r, h.cfg.SecureCookies),
			gs.ShowNamesPublicly,
			successMsg,
		).Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, "/settings?toast="+successMsg, http.StatusSeeOther)
}

// handleSettingsTokenRegenerateSubmit regenerates the public link token.
func (h *authHandlers) handleSettingsTokenRegenerateSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	newToken, err := settings.RegeneratePublicToken(h.conn)
	if err != nil {
		http.Error(w, "failed to regenerate token", http.StatusInternalServerError)
		return
	}

	successMsg := "Public link regenerated."
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		pages.SettingsPublicLink(
			admin.LanguagePref,
			auth.CSRFToken(w, r, h.cfg.SecureCookies),
			newToken,
			getBaseUrl(r),
			successMsg,
		).Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, "/settings?toast="+successMsg, http.StatusSeeOther)
}

// handleSettingsInfoSubmit updates group/masjid info.
func (h *authHandlers) handleSettingsInfoSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	gs.GroupName = strings.TrimSpace(r.FormValue("group_name"))
	gs.CurrencyCode = strings.TrimSpace(r.FormValue("currency_code"))
	gs.CurrencySymbol = strings.TrimSpace(r.FormValue("currency_symbol"))

	// Parse quick amounts
	quickAmountsRaw := r.FormValue("quick_amounts")
	parts := strings.Split(quickAmountsRaw, ",")
	var parsedAmounts []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if val, err := strconv.ParseInt(p, 10, 64); err == nil && val > 0 {
			parsedAmounts = append(parsedAmounts, val*100) // convert to minor units
		}
	}
	if len(parsedAmounts) > 0 {
		gs.QuickAmountsMinor = parsedAmounts
	}

	// Privacy policy URL
	privacyPolicyUrl := strings.TrimSpace(r.FormValue("privacy_policy_url"))
	if privacyPolicyUrl == "" {
		gs.PrivacyPolicyURL = nil
	} else {
		gs.PrivacyPolicyURL = &privacyPolicyUrl
	}

	if err := settings.Update(h.conn, gs); err != nil {
		http.Error(w, "failed to update settings", http.StatusInternalServerError)
		return
	}

	successMsg := "Group info saved."
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		pages.SettingsGroupInfo(
			admin.LanguagePref,
			auth.CSRFToken(w, r, h.cfg.SecureCookies),
			gs,
			successMsg,
		).Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, "/settings?toast="+successMsg, http.StatusSeeOther)
}

// handleSettingsLanguageSubmit updates the default public view language.
func (h *authHandlers) handleSettingsLanguageSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lang := r.FormValue("default_public_language")
	if lang == "en" || lang == "hi" || lang == "ar" {
		gs.DefaultPublicLanguage = lang
	}

	if err := settings.Update(h.conn, gs); err != nil {
		http.Error(w, "failed to update settings", http.StatusInternalServerError)
		return
	}

	successMsg := "Language settings saved."
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		pages.SettingsLanguage(
			admin.LanguagePref,
			auth.CSRFToken(w, r, h.cfg.SecureCookies),
			gs.DefaultPublicLanguage,
			successMsg,
		).Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, "/settings?toast="+successMsg, http.StatusSeeOther)
}
