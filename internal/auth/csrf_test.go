package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRFRoundTrip(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	token := CSRFToken(rec, req, true)
	if token == "" {
		t.Fatal("CSRFToken returned empty string")
	}

	// Simulate the browser sending the cookie back with the form post.
	form := url.Values{"csrf_token": {token}}
	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range rec.Result().Cookies() {
		postReq.AddCookie(c)
	}
	postReq.ParseForm()

	if err := VerifyCSRF(postReq); err != nil {
		t.Errorf("VerifyCSRF with matching cookie+field: %v", err)
	}
}

func TestCSRFRejectsMismatch(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	CSRFToken(rec, req, true)

	form := url.Values{"csrf_token": {"attacker-supplied-token"}}
	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range rec.Result().Cookies() {
		postReq.AddCookie(c)
	}
	postReq.ParseForm()

	if err := VerifyCSRF(postReq); err != ErrCSRFInvalid {
		t.Errorf("VerifyCSRF(mismatched) = %v, want ErrCSRFInvalid", err)
	}
}

func TestCSRFRejectsMissingCookie(t *testing.T) {
	form := url.Values{"csrf_token": {"some-token"}}
	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.ParseForm()

	if err := VerifyCSRF(postReq); err != ErrCSRFInvalid {
		t.Errorf("VerifyCSRF(no cookie) = %v, want ErrCSRFInvalid", err)
	}
}
