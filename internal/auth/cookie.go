package auth

import (
	"net/http"
	"time"
)

// SessionCookieName is the cookie holding the raw session token
// (docs/SCHEMA.md §5). The hash, never the raw value, is what's stored
// server-side.
const SessionCookieName = "sadqa_session"

// SetSessionCookie sets the session cookie: HttpOnly, SameSite=Lax always;
// Secure is configurable off only for local HTTP dev (docs/SCHEMA.md §5).
func SetSessionCookie(w http.ResponseWriter, rawToken string, lifetime time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(lifetime.Seconds()),
	})
}

// ClearSessionCookie expires the session cookie immediately.
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
