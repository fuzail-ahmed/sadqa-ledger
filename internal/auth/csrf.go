package auth

import (
	"errors"
	"net/http"
)

// csrfCookieName holds the double-submit CSRF token. It's HttpOnly since
// the server (not client JS) both sets the cookie and embeds the same value
// as a hidden form field when rendering a page — no client-side read needed.
const csrfCookieName = "csrf_token"

// ErrCSRFInvalid is returned by VerifyCSRF when the submitted form token
// doesn't match the cookie (missing, forged, or cross-site request).
var ErrCSRFInvalid = errors.New("invalid or missing CSRF token")

// CSRFToken returns the current request's CSRF token, setting a fresh one
// on the response if none exists yet. Call this when rendering any form
// that will POST to a state-changing route, and embed the returned value as
// a hidden field (docs/TRD.md §9).
func CSRFToken(w http.ResponseWriter, r *http.Request, secure bool) string {
	if c, err := r.Cookie(csrfCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	token, err := newRawToken()
	if err != nil {
		// crypto/rand failure is not recoverable; an empty token simply
		// fails every subsequent VerifyCSRF call, which is the safe default.
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

// VerifyCSRF checks the request's "csrf_token" form field against the
// csrf_token cookie (double-submit pattern). Call after r.ParseForm() in
// every POST/PUT/DELETE handler.
func VerifyCSRF(r *http.Request) error {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return ErrCSRFInvalid
	}
	formToken := r.FormValue("csrf_token")
	if formToken == "" || !constantTimeEqual(formToken, cookie.Value) {
		return ErrCSRFInvalid
	}
	return nil
}
