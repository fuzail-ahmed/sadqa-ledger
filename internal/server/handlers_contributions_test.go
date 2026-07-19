package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestContributionLoggingFlow(t *testing.T) {
	h, conn := newTestServer(t)
	completeSetup(t, h)
	loginGet := doGet(h, "/login", nil)
	loginCsrf, cookies := extractCSRF(t, loginGet)
	loginPost := doPostForm(h, "/login", url.Values{
		"csrf_token": {loginCsrf}, "username": {"sohail"}, "password": {"correct-horse-battery"},
	}, cookies)
	sessionCookies := mergeCookies(cookies, loginPost.Result().Cookies())

	// Create a member first
	newGet := doGet(h, "/members/new", sessionCookies)
	newCsrf, newCookies := extractCSRF(t, newGet)
	doPostForm(h, "/members/new", url.Values{
		"csrf_token": {newCsrf}, "name": {"Farhan"}, "is_active": {"1"},
	}, mergeCookies(sessionCookies, newCookies))

	var memberID int64
	conn.QueryRow(`SELECT id FROM members WHERE name = 'Farhan'`).Scan(&memberID)
	memberIDStr := strconv.FormatInt(memberID, 10)

	// --- 1. Get contribution new page ---
	contribPageGet := doGet(h, "/contributions/new", sessionCookies)
	if contribPageGet.Code != 200 {
		t.Fatalf("GET /contributions/new = %d, want 200", contribPageGet.Code)
	}
	contribCsrf, contribCookies := extractCSRF(t, contribPageGet)
	contribSessionCookies := mergeCookies(sessionCookies, contribCookies)

	// --- 2. Live search members ---
	reqSearch := httptest.NewRequest(http.MethodGet, "/contributions/search-members?member_search=Far", nil)
	for _, c := range contribSessionCookies {
		reqSearch.AddCookie(c)
	}
	recSearch := httptest.NewRecorder()
	h.ServeHTTP(recSearch, reqSearch)
	if recSearch.Code != 200 {
		t.Errorf("GET /contributions/search-members = %d, want 200", recSearch.Code)
	}
	if !strings.Contains(recSearch.Body.String(), "Farhan") {
		t.Errorf("Search results did not contain member 'Farhan'")
	}

	// --- 3. Check duplicates ---
	reqDupEmpty := httptest.NewRequest(http.MethodGet, "/contributions/check-duplicate?member_id="+memberIDStr+"&contribution_month=2026-07", nil)
	for _, c := range contribSessionCookies {
		reqDupEmpty.AddCookie(c)
	}
	recDupEmpty := httptest.NewRecorder()
	h.ServeHTTP(recDupEmpty, reqDupEmpty)
	if strings.Contains(recDupEmpty.Body.String(), "already has a payment logged") {
		t.Error("Empty database returned duplicate warning")
	}

	// --- 4. Submit invalid contribution (missing member) ---
	invalidPost := doPostForm(h, "/contributions/new", url.Values{
		"csrf_token":         {contribCsrf},
		"member_id":          {""},
		"amount":             {"500"},
		"contribution_month": {"2026-07"},
		"paid_on":            {"2026-07-19"},
	}, contribSessionCookies)
	if !strings.Contains(invalidPost.Body.String(), "Select a member.") {
		t.Error("Validation error for missing member not rendered")
	}

	// --- 5. Submit valid contribution ---
	validPost := doPostForm(h, "/contributions/new", url.Values{
		"csrf_token":         {contribCsrf},
		"member_id":          {memberIDStr},
		"amount":             {"500"},
		"contribution_month": {"2026-07"},
		"paid_on":            {"2026-07-19"},
	}, contribSessionCookies)
	if validPost.Code != 303 {
		t.Fatalf("POST /contributions/new = %d, want 303 redirect", validPost.Code)
	}

	// Verify database record
	var contribCount int
	var amountMinor int64
	var recordedByAdminID int64
	err := conn.QueryRow(`SELECT COUNT(*), amount_minor, recorded_by_admin_id FROM contributions WHERE member_id = ?`, memberID).Scan(&contribCount, &amountMinor, &recordedByAdminID)
	if err != nil {
		t.Fatalf("Query contribution: %v", err)
	}
	if contribCount != 1 {
		t.Errorf("contribCount = %d, want 1", contribCount)
	}
	if amountMinor != 50000 {
		t.Errorf("amountMinor = %d, want 50000 (₹500 in minor units)", amountMinor)
	}
	if recordedByAdminID != 1 {
		t.Errorf("recordedByAdminID = %d, want 1 (admin Sohail)", recordedByAdminID)
	}

	// --- 6. Check duplicate warning now appears ---
	reqDupActive := httptest.NewRequest(http.MethodGet, "/contributions/check-duplicate?member_id="+memberIDStr+"&contribution_month=2026-07", nil)
	for _, c := range contribSessionCookies {
		reqDupActive.AddCookie(c)
	}
	recDupActive := httptest.NewRecorder()
	h.ServeHTTP(recDupActive, reqDupActive)
	if !strings.Contains(recDupActive.Body.String(), "already has a payment logged") {
		t.Error("Duplicate check did not return warning after contribution creation")
	}

	// --- 7. Submit valid contribution via HTMX ---
	reqHtmx := httptest.NewRequest(http.MethodPost, "/contributions/new", strings.NewReader(url.Values{
		"csrf_token":         {contribCsrf},
		"member_id":          {memberIDStr},
		"amount":             {"200"},
		"contribution_month": {"2026-07"},
		"paid_on":            {"2026-07-19"},
	}.Encode()))
	reqHtmx.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqHtmx.Header.Set("HX-Request", "true")
	for _, c := range contribSessionCookies {
		reqHtmx.AddCookie(c)
	}
	recHtmx := httptest.NewRecorder()
	h.ServeHTTP(recHtmx, reqHtmx)
	if recHtmx.Code != 200 {
		t.Fatalf("HTMX POST /contributions/new = %d, want 200", recHtmx.Code)
	}
	htmxBody := recHtmx.Body.String()
	if !strings.Contains(htmxBody, "Saved —") {
		t.Error("HTMX response did not return success toast message")
	}
	if !strings.Contains(htmxBody, `id="amount"`) || !strings.Contains(htmxBody, `value=""`) {
		t.Errorf("HTMX response did not reset amount field. Body was:\n%s", htmxBody)
	}
}
