package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestEndToEndAdminSession(t *testing.T) {
	h, conn := newTestServer(t)

	// 1. Run Setup Wizard
	setupGet := doGet(h, "/setup", nil)
	setupCsrf, setupCookies := extractCSRF(t, setupGet)

	setupReq := doPostForm(h, "/setup", url.Values{
		"csrf_token":      {setupCsrf},
		"group_name":      {"Masjid Bilal"},
		"currency_code":   {"INR"},
		"currency_symbol": {"₹"},
		"username":        {"sohail"},
		"display_name":    {"Sohail Admin"},
		"password":        {"correct-horse-battery-staple"},
	}, setupCookies)
	if setupReq.Code != 303 {
		t.Fatalf("POST /setup = %d, want 303", setupReq.Code)
	}

	// 2. Perform Login
	loginGet := doGet(h, "/login", nil)
	loginCsrf, cookies := extractCSRF(t, loginGet)

	loginPost := doPostForm(h, "/login", url.Values{
		"csrf_token": {loginCsrf},
		"username":   {"sohail"},
		"password":   {"correct-horse-battery-staple"},
	}, cookies)
	if loginPost.Code != 303 {
		t.Fatalf("POST /login = %d, want 303 redirect", loginPost.Code)
	}
	sessionCookies := mergeCookies(cookies, loginPost.Result().Cookies())

	// 3. Add Member
	newMemberGet := doGet(h, "/members/new", sessionCookies)
	newMemberCsrf, newMemberCookies := extractCSRF(t, newMemberGet)
	memberSessionCookies := mergeCookies(sessionCookies, newMemberCookies)

	addMemberPost := doPostForm(h, "/members/new", url.Values{
		"csrf_token": {newMemberCsrf},
		"name":       {"Arsalan Tester"},
		"is_active":  {"1"},
	}, memberSessionCookies)
	if addMemberPost.Code != 303 {
		t.Fatalf("POST /members/new = %d, want 303 redirect", addMemberPost.Code)
	}

	var memberID int64
	err := conn.QueryRow(`SELECT id FROM members WHERE name = 'Arsalan Tester'`).Scan(&memberID)
	if err != nil {
		t.Fatalf("Failed to find added member in DB: %v", err)
	}

	// 4. Verify Checklist shows Arsalan as UNPAID
	currentMonth := time.Now().Format("2006-01")
	dashboardGet := doGet(h, "/", sessionCookies)
	dashboardBody := dashboardGet.Body.String()
	if !strings.Contains(dashboardBody, "Arsalan Tester") {
		t.Fatal("Arsalan Tester not present in checklist")
	}
	if !strings.Contains(dashboardBody, `aria-label="Unpaid"`) {
		t.Error("Arsalan is not marked as Unpaid initially")
	}

	// 5. Log Contribution
	contribGet := doGet(h, "/contributions/new", sessionCookies)
	contribCsrf, contribCookies := extractCSRF(t, contribGet)
	contribSessionCookies := mergeCookies(sessionCookies, contribCookies)

	addContribPost := doPostForm(h, "/contributions/new", url.Values{
		"csrf_token":         {contribCsrf},
		"member_id":          {fmt.Sprintf("%d", memberID)},
		"amount":             {"1000"}, // 1000 rupees
		"contribution_month": {currentMonth},
		"paid_on":            {time.Now().Format("2006-01-02")},
	}, contribSessionCookies)
	if addContribPost.Code != 303 {
		t.Fatalf("POST /contributions/new = %d, want 303 redirect", addContribPost.Code)
	}

	// 6. Verify Checklist updates to PAID
	dashboardAfterGet := doGet(h, "/", sessionCookies)
	dashboardAfterBody := dashboardAfterGet.Body.String()
	if !strings.Contains(dashboardAfterBody, `✓`) {
		t.Error("Arsalan did not update to Paid (checkmark missing)")
	}
	if !strings.Contains(dashboardAfterBody, `₹1000`) {
		t.Error("Dashboard stats/checklist amount ₹1000 missing")
	}

	// 7. Verify Dashboard totals
	if !strings.Contains(dashboardAfterBody, `₹1000`) {
		t.Error("All-time total collected stat mismatch")
	}

	// 8. Turn public names visibility ON to verify it in the WhatsApp summary
	settingsGet := doGet(h, "/settings", sessionCookies)
	settingsCsrf, settingsCookies := extractCSRF(t, settingsGet)
	settingsSessionCookies := mergeCookies(sessionCookies, settingsCookies)

	privacyUpdateRes := doPostForm(h, "/settings/privacy", url.Values{
		"csrf_token":          {settingsCsrf},
		"show_names_publicly": {"1"},
	}, settingsSessionCookies)
	if privacyUpdateRes.Code != 303 {
		t.Fatalf("POST /settings/privacy = %d, want 303", privacyUpdateRes.Code)
	}

	// 9. Generate Monthly Summary Text
	reqSummary := httptest.NewRequest(http.MethodGet, "/summary/generate?summary_month="+currentMonth, nil)
	for _, c := range sessionCookies {
		reqSummary.AddCookie(c)
	}
	recSummary := httptest.NewRecorder()
	h.ServeHTTP(recSummary, reqSummary)
	if recSummary.Code != 200 {
		t.Fatalf("GET /summary/generate = %d, want 200", recSummary.Code)
	}

	summaryText := recSummary.Body.String()
	if !strings.Contains(summaryText, "*Sadqa Ledger") {
		t.Error("Generated summary formatting is incorrect")
	}
	if !strings.Contains(summaryText, "Total Collected: ₹1000") {
		t.Error("Generated summary missing total collected amount")
	}
	if !strings.Contains(summaryText, "Arsalan Tester: ₹1000") {
		t.Error("Generated summary missing Arsalan Tester contribution details when names are public")
	}
}
