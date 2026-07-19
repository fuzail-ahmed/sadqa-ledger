package server

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/fuzail-ahmed/sadqa-ledger/i18n"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/config"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

const minPasswordLength = 8

// authHandlers holds the dependencies shared by every auth route: the DB
// connection and the config values that shape cookies/session lifetime.
type authHandlers struct {
	conn *sql.DB
	cfg  config.Config
}

// handleSetupPage renders the first-run wizard, or redirects to /login if
// setup has already been completed (docs/IMPLEMENTATION_PLAN.md Phase 2
// acceptance criterion: setup is unreachable once an admin exists).
func (h *authHandlers) handleSetupPage(w http.ResponseWriter, r *http.Request) {
	done, err := auth.GroupSettingsExist(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if done {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	csrfToken := auth.CSRFToken(w, r, h.cfg.SecureCookies)
	pages.Setup(csrfToken, "", "", "", pages.SetupErrors{}, "").Render(r.Context(), w)
}

// handleSetupSubmit creates the first admin and the group_settings row
// together (docs/SCHEMA.md §8), then logs the new admin straight in.
func (h *authHandlers) handleSetupSubmit(w http.ResponseWriter, r *http.Request) {
	done, err := auth.GroupSettingsExist(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if done {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	groupName := strings.TrimSpace(r.FormValue("group_name"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	var errs pages.SetupErrors
	if groupName == "" {
		errs.GroupName = i18n.T("en", "validation.required")
	}
	if displayName == "" {
		errs.DisplayName = i18n.T("en", "validation.required")
	}
	if username == "" {
		errs.Username = i18n.T("en", "validation.required")
	}
	if len(password) < minPasswordLength {
		errs.Password = i18n.T("en", "validation.password_too_short")
	}
	if errs != (pages.SetupErrors{}) {
		csrfToken := auth.CSRFToken(w, r, h.cfg.SecureCookies)
		pages.Setup(csrfToken, groupName, displayName, username, errs, "").Render(r.Context(), w)
		return
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	adminID, err := auth.CreateGroupInSetup(h.conn, username, passwordHash, displayName, groupName, h.cfg.DefaultCurrencyCode, h.cfg.DefaultCurrencySymbol)
	if err != nil {
		csrfToken := auth.CSRFToken(w, r, h.cfg.SecureCookies)
		errs.Username = i18n.T("en", "validation.username_taken")
		pages.Setup(csrfToken, groupName, displayName, username, errs, "").Render(r.Context(), w)
		return
	}

	h.startSession(w, r, adminID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleLoginPage renders the login form. Already-authenticated admins are
// redirected to /, and if setup hasn't run yet the wizard takes priority
// over the login form (docs/APP_FLOW.md §1).
func (h *authHandlers) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	done, err := auth.GroupSettingsExist(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !done {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}
	if h.currentSession(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	csrfToken := auth.CSRFToken(w, r, h.cfg.SecureCookies)
	pages.Login(csrfToken, "", "").Render(r.Context(), w)
}

// handleLoginSubmit authenticates the submitted credentials and, on
// success, regenerates a fresh session (no fixation, docs/SCHEMA.md §5).
// The error message and timing are identical for an unknown username and a
// wrong password, to avoid username enumeration (docs/APP_FLOW.md §1).
func (h *authHandlers) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	admin, err := auth.FindActiveAdminByUsername(h.conn, username)
	if err != nil || !auth.VerifyPassword(admin.PasswordHash, password) {
		if err != nil {
			auth.VerifyAgainstDummy(password)
		}
		csrfToken := auth.CSRFToken(w, r, h.cfg.SecureCookies)
		pages.Login(csrfToken, username, i18n.T("en", "login.error")).Render(r.Context(), w)
		return
	}

	h.startSession(w, r, admin.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleLogout deletes the server-side session row and clears the cookie
// (docs/SCHEMA.md §5 — logout is not just a cookie clear).
func (h *authHandlers) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
		auth.DeleteSession(h.conn, cookie.Value)
	}
	auth.ClearSessionCookie(w, h.cfg.SecureCookies)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// handleAdminNewPage renders the add-admin form (existing admins only, wired
// behind RequireAuth in server.go). The full Settings > Admins section
// (list, remove, reset-password) is a Phase 7 deliverable; this ships just
// enough to satisfy "existing admins can create additional admins" now.
func (h *authHandlers) handleAdminNewPage(w http.ResponseWriter, r *http.Request) {
	pages.AdminNew(h.shellData(w, r, "", "admin_new.heading"), "", "", pages.AdminNewErrors{}, "").Render(r.Context(), w)
}

func (h *authHandlers) handleAdminNewSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	creator := auth.CurrentAdmin(r)
	lang := creator.LanguagePref

	username := strings.TrimSpace(r.FormValue("username"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	password := r.FormValue("password")

	var errs pages.AdminNewErrors
	if username == "" {
		errs.Username = i18n.T(lang, "validation.required")
	}
	if displayName == "" {
		errs.DisplayName = i18n.T(lang, "validation.required")
	}
	if len(password) < minPasswordLength {
		errs.Password = i18n.T(lang, "validation.password_too_short")
	}
	if errs != (pages.AdminNewErrors{}) {
		pages.AdminNew(h.shellData(w, r, "", "admin_new.heading"), username, displayName, errs, "").Render(r.Context(), w)
		return
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, err := auth.CreateAdmin(h.conn, username, passwordHash, displayName, creator.ID); err != nil {
		errs.Username = i18n.T(lang, "validation.username_taken")
		pages.AdminNew(h.shellData(w, r, "", "admin_new.heading"), username, displayName, errs, "").Render(r.Context(), w)
		return
	}

	pages.AdminNew(h.shellData(w, r, "", "admin_new.heading"), "", "", pages.AdminNewErrors{}, i18n.T(lang, "admin_new.success")).Render(r.Context(), w)
}

// startSession creates a fresh session and sets the cookie — the single
// path both /login and /setup use to log an admin in.
func (h *authHandlers) startSession(w http.ResponseWriter, r *http.Request, adminID int64) {
	rawToken, err := auth.CreateSession(h.conn, adminID, h.cfg.SessionLifetime, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	auth.SetSessionCookie(w, rawToken, h.cfg.SessionLifetime, h.cfg.SecureCookies)
}

// currentSession returns the looked-up session for the request's cookie, or
// nil if there isn't a valid one — used only to decide whether /login
// should redirect an already-authenticated admin straight to /.
func (h *authHandlers) currentSession(r *http.Request) *auth.Session {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err != nil {
		return nil
	}
	sess, err := auth.LookupSession(h.conn, cookie.Value)
	if err != nil {
		return nil
	}
	return sess
}
