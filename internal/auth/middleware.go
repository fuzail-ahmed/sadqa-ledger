package auth

import (
	"context"
	"database/sql"
	"net/http"
)

type contextKey int

const adminContextKey contextKey = iota

// CurrentAdmin returns the authenticated admin attached to the request by
// RequireAuth, or nil if the route isn't behind that middleware.
func CurrentAdmin(r *http.Request) *Admin {
	admin, _ := r.Context().Value(adminContextKey).(*Admin)
	return admin
}

// redirectToLogin sends the browser to /login and clears any stale session
// cookie, per the lookup-flow failure case in docs/SCHEMA.md §5.
func redirectToLogin(w http.ResponseWriter, r *http.Request, secure bool) {
	ClearSessionCookie(w, secure)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// RequireAuth is chi middleware protecting admin routes (docs/SCHEMA.md §6:
// enforcement at the HTTP-handler layer). An empty database (no
// group_settings row yet) redirects to /setup instead of /login, per
// docs/IMPLEMENTATION_PLAN.md Phase 2's acceptance criteria. A missing,
// invalid, or expired session redirects to /login.
func RequireAuth(conn *sql.DB, secureCookie bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setupDone, err := GroupSettingsExist(conn)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !setupDone {
				http.Redirect(w, r, "/setup", http.StatusSeeOther)
				return
			}

			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				redirectToLogin(w, r, secureCookie)
				return
			}

			sess, err := LookupSession(conn, cookie.Value)
			if err != nil {
				redirectToLogin(w, r, secureCookie)
				return
			}

			admin, err := FindAdminByID(conn, sess.AdminID)
			if err != nil {
				redirectToLogin(w, r, secureCookie)
				return
			}

			ctx := context.WithValue(r.Context(), adminContextKey, admin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
