package server

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/config"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/db"
	"github.com/fuzail-ahmed/sadqa-ledger/migrations"
)

func newTestServer(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()
	conn, err := db.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := db.Migrate(conn, migrations.FS); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	cfg := config.Config{
		SecureCookies:         false, // http test server
		SessionLifetime:       time.Hour,
		DefaultCurrencyCode:   "INR",
		DefaultCurrencySymbol: "₹",
	}
	return New(conn, cfg), conn
}

var csrfInputRe = regexp.MustCompile(`name="csrf_token" value="([^"]+)"`)

// extractCSRF pulls the hidden csrf_token field's value out of a rendered
// page and returns it along with the response's cookies, which must be
// replayed on the follow-up POST (double-submit pattern).
func extractCSRF(t *testing.T, rec *httptest.ResponseRecorder) (string, []*http.Cookie) {
	t.Helper()
	body, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	m := csrfInputRe.FindStringSubmatch(string(body))
	if m == nil {
		t.Fatalf("csrf_token field not found in body:\n%s", body)
	}
	return m[1], rec.Result().Cookies()
}

func doGet(h http.Handler, path string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func doPostForm(h http.Handler, path string, form url.Values, cookies []*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func mergeCookies(sets ...[]*http.Cookie) []*http.Cookie {
	var out []*http.Cookie
	seen := map[string]bool{}
	for i := len(sets) - 1; i >= 0; i-- {
		for _, c := range sets[i] {
			if !seen[c.Name] {
				out = append(out, c)
				seen[c.Name] = true
			}
		}
	}
	return out
}

func TestProtectedRouteRedirectsToSetupOnEmptyDB(t *testing.T) {
	h, _ := newTestServer(t)
	rec := doGet(h, "/", nil)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/setup" {
		t.Errorf("GET / on empty DB = %d %q, want 303 /setup", rec.Code, rec.Header().Get("Location"))
	}
}

func TestHealthzIsPublicAndPingsDatabase(t *testing.T) {
	h, _ := newTestServer(t)
	rec := doGet(h, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok\n" {
		t.Fatalf("GET /healthz body = %q, want ok", rec.Body.String())
	}
}

func TestFullSignupLoginLogoutFlow(t *testing.T) {
	h, conn := newTestServer(t)

	// --- Setup ---
	setupGet := doGet(h, "/setup", nil)
	csrfTok, cookies := extractCSRF(t, setupGet)

	setupPost := doPostForm(h, "/setup", url.Values{
		"csrf_token":   {csrfTok},
		"group_name":   {"Test Masjid"},
		"display_name": {"Sohail"},
		"username":     {"sohail"},
		"password":     {"correct-horse-battery"},
	}, cookies)
	if setupPost.Code != http.StatusSeeOther || setupPost.Header().Get("Location") != "/" {
		t.Fatalf("POST /setup = %d %q, want 303 /", setupPost.Code, setupPost.Header().Get("Location"))
	}
	sessionCookies := mergeCookies(cookies, setupPost.Result().Cookies())

	// --- Setup now unreachable ---
	setupAgain := doGet(h, "/setup", nil)
	if setupAgain.Code != http.StatusSeeOther || setupAgain.Header().Get("Location") != "/login" {
		t.Errorf("GET /setup after admin exists = %d %q, want 303 /login", setupAgain.Code, setupAgain.Header().Get("Location"))
	}

	// --- Protected page reachable with the session from setup ---
	home := doGet(h, "/", sessionCookies)
	if home.Code != http.StatusOK {
		t.Fatalf("GET / with valid session = %d, want 200", home.Code)
	}
	if !strings.Contains(home.Body.String(), "Sohail") {
		t.Error("home page doesn't show the logged-in admin's display name")
	}

	// --- Logout ---
	logoutCsrf, logoutCookies := extractCSRF(t, home)
	logoutPost := doPostForm(h, "/logout", url.Values{"csrf_token": {logoutCsrf}}, mergeCookies(sessionCookies, logoutCookies))
	if logoutPost.Code != http.StatusSeeOther || logoutPost.Header().Get("Location") != "/login" {
		t.Fatalf("POST /logout = %d %q, want 303 /login", logoutPost.Code, logoutPost.Header().Get("Location"))
	}

	// --- Protected page redirects again after logout ---
	afterLogout := doGet(h, "/", sessionCookies)
	if afterLogout.Code != http.StatusSeeOther || afterLogout.Header().Get("Location") != "/login" {
		t.Errorf("GET / after logout = %d %q, want 303 /login", afterLogout.Code, afterLogout.Header().Get("Location"))
	}

	// --- DB never contains the raw session token ---
	var rawTokenLeaked int
	for _, c := range sessionCookies {
		if c.Name == "sadqa_session" {
			conn.QueryRow(`SELECT COUNT(*) FROM sessions WHERE token_hash = ?`, c.Value).Scan(&rawTokenLeaked)
		}
	}
	if rawTokenLeaked != 0 {
		t.Error("raw session token cookie value found stored directly in sessions table")
	}
}

func TestLoginWrongPasswordAndUnknownUsernameGiveSameError(t *testing.T) {
	h, _ := newTestServer(t)
	completeSetup(t, h)

	loginGet := doGet(h, "/login", nil)
	csrfTok, cookies := extractCSRF(t, loginGet)

	wrongPass := doPostForm(h, "/login", url.Values{
		"csrf_token": {csrfTok}, "username": {"sohail"}, "password": {"wrong"},
	}, cookies)
	unknownUser := doPostForm(h, "/login", url.Values{
		"csrf_token": {csrfTok}, "username": {"ghost"}, "password": {"wrong"},
	}, cookies)

	if wrongPass.Code != http.StatusOK || unknownUser.Code != http.StatusOK {
		t.Fatalf("expected both failed logins to re-render 200, got %d and %d", wrongPass.Code, unknownUser.Code)
	}
	wrongPassBody := wrongPass.Body.String()
	unknownUserBody := unknownUser.Body.String()
	if !strings.Contains(wrongPassBody, "Incorrect username or password") {
		t.Error("wrong-password response missing generic error text")
	}
	if !strings.Contains(unknownUserBody, "Incorrect username or password") {
		t.Error("unknown-username response missing generic error text")
	}
}

func TestCSRFRejectedOnStateChangingRoute(t *testing.T) {
	h, _ := newTestServer(t)
	loginGet := doGet(h, "/setup", nil)
	_, cookies := extractCSRF(t, loginGet)

	rec := doPostForm(h, "/setup", url.Values{
		"csrf_token":   {"forged-token"},
		"group_name":   {"Test Masjid"},
		"display_name": {"Sohail"},
		"username":     {"sohail"},
		"password":     {"correct-horse-battery"},
	}, cookies)
	if rec.Code != http.StatusForbidden {
		t.Errorf("POST /setup with forged CSRF token = %d, want 403", rec.Code)
	}
}

// completeSetup runs the setup POST for a standard admin
// (sohail / correct-horse-battery), used by tests that need setup already done.
func completeSetup(t *testing.T, h http.Handler) {
	t.Helper()
	setupGet := doGet(h, "/setup", nil)
	csrfTok, cookies := extractCSRF(t, setupGet)
	doPostForm(h, "/setup", url.Values{
		"csrf_token":   {csrfTok},
		"group_name":   {"Test Masjid"},
		"display_name": {"Sohail"},
		"username":     {"sohail"},
		"password":     {"correct-horse-battery"},
	}, cookies)
}
