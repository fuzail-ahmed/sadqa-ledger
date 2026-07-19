package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/settings"
)

func TestPublicAndSummarySettingsFlow(t *testing.T) {
	h, conn := newTestServer(t)
	completeSetup(t, h)
	loginGet := doGet(h, "/login", nil)
	loginCsrf, cookies := extractCSRF(t, loginGet)
	loginPost := doPostForm(h, "/login", url.Values{
		"csrf_token": {loginCsrf}, "username": {"sohail"}, "password": {"correct-horse-battery"},
	}, cookies)
	sessionCookies := mergeCookies(cookies, loginPost.Result().Cookies())

	// Create a member and a contribution for testing visibility
	newMemberGet := doGet(h, "/members/new", sessionCookies)
	newMemberCsrf, newMemberCookies := extractCSRF(t, newMemberGet)
	doPostForm(h, "/members/new", url.Values{
		"csrf_token": {newMemberCsrf}, "name": {"Farhan Member"}, "is_active": {"1"},
	}, mergeCookies(sessionCookies, newMemberCookies))

	var memberID int64
	conn.QueryRow(`SELECT id FROM members WHERE name = 'Farhan Member'`).Scan(&memberID)

	contribGet := doGet(h, "/contributions/new", sessionCookies)
	contribCsrf, contribCookies := extractCSRF(t, contribGet)
	doPostForm(h, "/contributions/new", url.Values{
		"csrf_token":         {contribCsrf},
		"member_id":          {fmt.Sprintf("%d", memberID)},
		"amount":             {"500"},
		"contribution_month": {"2026-07"},
		"paid_on":            {"2026-07-19"},
	}, mergeCookies(sessionCookies, contribCookies))

	// Get initial settings
	gs, err := settings.Get(conn)
	if err != nil {
		t.Fatalf("Get settings: %v", err)
	}

	// --- 1. Public Transparency Page (default: names hidden) ---
	publicUrl := "/p/" + gs.PublicToken
	publicRes := doGet(h, publicUrl, nil)
	if publicRes.Code != 200 {
		t.Fatalf("GET %s = %d, want 200", publicUrl, publicRes.Code)
	}
	publicBody := publicRes.Body.String()
	if !strings.Contains(publicBody, "Public Transparency Ledger") {
		t.Error("Public page title/text missing")
	}
	// Check noindex header
	if robots := publicRes.Header().Get("X-Robots-Tag"); robots != "noindex, nofollow" {
		t.Errorf("X-Robots-Tag = %q, want 'noindex, nofollow'", robots)
	}
	// Check that names are hidden (show_names_publicly is false by default)
	if strings.Contains(publicBody, "Farhan Member") {
		t.Error("Public page contains member name Farhan Member, but privacy setting is OFF")
	}

	// --- 2. Invalid Token returns generic 404 ---
	invalidRes := doGet(h, "/p/badtoken123", nil)
	if invalidRes.Code != 404 {
		t.Errorf("GET /p/badtoken123 = %d, want 404", invalidRes.Code)
	}
	if !strings.Contains(invalidRes.Body.String(), "This page isn't available.") {
		t.Error("Invalid token page did not render generic unavailable message")
	}

	// --- 3. WhatsApp Summary (default: names hidden) ---
	summaryRes := doGet(h, "/summary", sessionCookies)
	if summaryRes.Code != 200 {
		t.Fatalf("GET /summary = %d, want 200", summaryRes.Code)
	}
	summaryBody := summaryRes.Body.String()
	if strings.Contains(summaryBody, "Farhan Member") {
		t.Error("Summary contains member name Farhan Member, but privacy setting is OFF")
	}

	// --- 4. Update Settings (Turn show_names_publicly ON) ---
	settingsGet := doGet(h, "/settings", sessionCookies)
	settingsCsrf, settingsCookies := extractCSRF(t, settingsGet)
	settingsSessionCookies := mergeCookies(sessionCookies, settingsCookies)

	privacyUpdateRes := doPostForm(h, "/settings/privacy", url.Values{
		"csrf_token":          {settingsCsrf},
		"show_names_publicly": {"1"},
	}, settingsSessionCookies)
	if privacyUpdateRes.Code != 303 {
		t.Errorf("POST /settings/privacy = %d, want 303 redirect", privacyUpdateRes.Code)
	}

	// Verify settings updated
	gsUpdated, _ := settings.Get(conn)
	if !gsUpdated.ShowNamesPublicly {
		t.Error("show_names_publicly not updated to true")
	}

	// --- 5. Public page (names visible) ---
	publicResVisible := doGet(h, publicUrl, nil)
	publicVisibleBody := publicResVisible.Body.String()
	if !strings.Contains(publicVisibleBody, "Farhan Member") {
		t.Error("Public page missing member name Farhan Member, but privacy setting is ON")
	}

	// --- 6. Summary (names visible) ---
	reqSummaryGen := httptest.NewRequest(http.MethodGet, "/summary/generate?summary_month=2026-07", nil)
	for _, c := range settingsSessionCookies {
		reqSummaryGen.AddCookie(c)
	}
	recSummaryGen := httptest.NewRecorder()
	h.ServeHTTP(recSummaryGen, reqSummaryGen)
	if recSummaryGen.Code != 200 {
		t.Fatalf("GET /summary/generate = %d, want 200", recSummaryGen.Code)
	}
	summaryGenBody := recSummaryGen.Body.String()
	if !strings.Contains(summaryGenBody, "*Sadqa Ledger — July 2026*") {
		t.Error("Generated summary missing header")
	}
	if !strings.Contains(summaryGenBody, "Farhan Member: ₹500") {
		t.Error("Generated summary missing contributor line Farhan Member: ₹500 when privacy is ON")
	}

	// --- 7. Regenerate public token ---
	oldToken := gsUpdated.PublicToken
	regenRes := doPostForm(h, "/settings/regenerate-token", url.Values{
		"csrf_token": {settingsCsrf},
	}, settingsSessionCookies)
	if regenRes.Code != 303 {
		t.Fatalf("POST /settings/regenerate-token = %d, want 303", regenRes.Code)
	}

	gsRegen, _ := settings.Get(conn)
	if gsRegen.PublicToken == oldToken {
		t.Error("PublicToken did not change after regeneration")
	}

	// Old link must return 404 now
	oldLinkRes := doGet(h, "/p/"+oldToken, nil)
	if oldLinkRes.Code != 404 {
		t.Errorf("Old public token link = %d, want 404", oldLinkRes.Code)
	}

	// New link must return 200
	newLinkRes := doGet(h, "/p/"+gsRegen.PublicToken, nil)
	if newLinkRes.Code != 200 {
		t.Errorf("New public token link = %d, want 200", newLinkRes.Code)
	}

	// --- 8. Update Group Info ---
	infoRes := doPostForm(h, "/settings/info", url.Values{
		"csrf_token":         {settingsCsrf},
		"group_name":         {"Masjid Al-Noor"},
		"currency_code":      {"USD"},
		"currency_symbol":    {"$"},
		"quick_amounts":      {"10, 20, 50"},
		"privacy_policy_url": {"https://masjid.org/privacy"},
	}, settingsSessionCookies)
	if infoRes.Code != 303 {
		t.Errorf("POST /settings/info = %d, want 303", infoRes.Code)
	}

	gsFinal, _ := settings.Get(conn)
	if gsFinal.GroupName != "Masjid Al-Noor" {
		t.Errorf("gsFinal.GroupName = %q, want Masjid Al-Noor", gsFinal.GroupName)
	}
	if gsFinal.CurrencySymbol != "$" {
		t.Errorf("gsFinal.CurrencySymbol = %q, want $", gsFinal.CurrencySymbol)
	}
	if gsFinal.PrivacyPolicyURL == nil || *gsFinal.PrivacyPolicyURL != "https://masjid.org/privacy" {
		t.Errorf("gsFinal.PrivacyPolicyURL = %v, want 'https://masjid.org/privacy'", gsFinal.PrivacyPolicyURL)
	}
	if len(gsFinal.QuickAmountsMinor) != 3 || gsFinal.QuickAmountsMinor[0] != 1000 {
		t.Errorf("gsFinal.QuickAmountsMinor = %v, want [1000, 2000, 5000]", gsFinal.QuickAmountsMinor)
	}

	// --- 9. PWA Routes ---
	manifestRes := doGet(h, "/manifest.json", nil)
	if manifestRes.Code != 200 {
		t.Errorf("GET /manifest.json = %d, want 200", manifestRes.Code)
	}
	if !strings.Contains(manifestRes.Body.String(), "Sadqa Ledger") {
		t.Error("manifest.json missing name field")
	}

	swRes := doGet(h, "/sw.js", nil)
	if swRes.Code != 200 {
		t.Errorf("GET /sw.js = %d, want 200", swRes.Code)
	}
	if !strings.Contains(swRes.Body.String(), "sadqa-ledger-v1") {
		t.Error("sw.js missing cache name definition")
	}
}
