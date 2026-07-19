package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestMemberAddEditDeactivateFlow(t *testing.T) {
	h, conn := newTestServer(t)
	completeSetup(t, h)
	loginGet := doGet(h, "/login", nil)
	loginCsrf, cookies := extractCSRF(t, loginGet)
	loginPost := doPostForm(h, "/login", url.Values{
		"csrf_token": {loginCsrf}, "username": {"sohail"}, "password": {"correct-horse-battery"},
	}, cookies)
	sessionCookies := mergeCookies(cookies, loginPost.Result().Cookies())

	// --- Empty state ---
	listEmpty := doGet(h, "/members", sessionCookies)
	if listEmpty.Code != 200 {
		t.Fatalf("GET /members = %d, want 200", listEmpty.Code)
	}
	if !strings.Contains(listEmpty.Body.String(), "No members yet") {
		t.Error("empty member list missing empty-state text")
	}

	// --- Validation error on add ---
	newGet := doGet(h, "/members/new", sessionCookies)
	newCsrf, newCookies := extractCSRF(t, newGet)
	invalidAdd := doPostForm(h, "/members/new", url.Values{
		"csrf_token": {newCsrf}, "name": {""}, "is_active": {"1"},
	}, mergeCookies(sessionCookies, newCookies))
	if invalidAdd.Code != 200 || !strings.Contains(invalidAdd.Body.String(), "required") {
		t.Errorf("POST /members/new with blank name = %d, want 200 with a required-field error", invalidAdd.Code)
	}

	// --- Successful add ---
	validAdd := doPostForm(h, "/members/new", url.Values{
		"csrf_token": {newCsrf}, "name": {"Farhan"}, "is_active": {"1"},
	}, mergeCookies(sessionCookies, newCookies))
	if validAdd.Code != 303 || validAdd.Header().Get("Location") != "/members?toast=added" {
		t.Fatalf("POST /members/new = %d %q, want 303 /members?toast=added", validAdd.Code, validAdd.Header().Get("Location"))
	}

	var memberID int64
	if err := conn.QueryRow(`SELECT id FROM members WHERE name = 'Farhan'`).Scan(&memberID); err != nil {
		t.Fatalf("member not created: %v", err)
	}
	var createdBy int64
	conn.QueryRow(`SELECT created_by_admin_id FROM members WHERE id = ?`, memberID).Scan(&createdBy)
	if createdBy == 0 {
		t.Error("created_by_admin_id not recorded on member creation")
	}

	// --- List reflects the new member, and the toast shows ---
	listAfterAdd := doGet(h, "/members?toast=added", sessionCookies)
	body := listAfterAdd.Body.String()
	if !strings.Contains(body, "Farhan") {
		t.Error("member list doesn't show the newly added member")
	}
	if !strings.Contains(body, "Member added") {
		t.Error("member list missing the added-toast message")
	}

	// --- Search with no results ---
	noResults := doGet(h, "/members?q=zzz", sessionCookies)
	if !strings.Contains(noResults.Body.String(), "No members match") {
		t.Error("search-no-results state text missing")
	}

	// --- Edit ---
	idStr := strconv.FormatInt(memberID, 10)
	editPath := "/members/" + idStr + "/edit"
	editGet := doGet(h, editPath, sessionCookies)
	editCsrf, editCookies := extractCSRF(t, editGet)
	editPost := doPostForm(h, editPath, url.Values{
		"csrf_token": {editCsrf}, "name": {"Farhan Ahmed"}, "is_active": {"1"},
	}, mergeCookies(sessionCookies, editCookies))
	if editPost.Code != 303 || editPost.Header().Get("Location") != "/members?toast=updated" {
		t.Fatalf("POST %s = %d %q, want 303 /members?toast=updated", editPath, editPost.Code, editPost.Header().Get("Location"))
	}
	var updatedBy int64
	conn.QueryRow(`SELECT updated_by_admin_id FROM members WHERE id = ?`, memberID).Scan(&updatedBy)
	if updatedBy == 0 {
		t.Error("updated_by_admin_id not recorded on member edit")
	}

	// --- Deactivate (soft delete) ---
	togglePath := "/members/" + idStr + "/toggle"
	toggleGet := doGet(h, "/members", sessionCookies) // fresh CSRF token
	toggleCsrf, toggleCookies := extractCSRF(t, toggleGet)
	togglePost := doPostForm(h, togglePath, url.Values{"csrf_token": {toggleCsrf}}, mergeCookies(sessionCookies, toggleCookies))
	if togglePost.Code != 303 || togglePost.Header().Get("Location") != "/members" {
		t.Fatalf("POST %s = %d %q, want 303 /members", togglePath, togglePost.Code, togglePost.Header().Get("Location"))
	}

	var isActive int
	if err := conn.QueryRow(`SELECT is_active FROM members WHERE id = ?`, memberID).Scan(&isActive); err != nil {
		t.Fatalf("query member: %v", err)
	}
	if isActive != 0 {
		t.Error("member still active after deactivation toggle")
	}
	var stillExists int
	conn.QueryRow(`SELECT COUNT(*) FROM members WHERE id = ?`, memberID).Scan(&stillExists)
	if stillExists != 1 {
		t.Fatal("member row deleted instead of soft-deactivated")
	}

	// --- Reactivate ---
	togglePost2 := doPostForm(h, togglePath, url.Values{"csrf_token": {toggleCsrf}}, mergeCookies(sessionCookies, toggleCookies))
	if togglePost2.Code != 303 {
		t.Fatalf("second toggle (reactivate) = %d, want 303", togglePost2.Code)
	}
	conn.QueryRow(`SELECT is_active FROM members WHERE id = ?`, memberID).Scan(&isActive)
	if isActive != 1 {
		t.Error("member still inactive after reactivation toggle")
	}
}

func TestMembersLiveSearchFragmentOmitsShell(t *testing.T) {
	h, _ := newTestServer(t)
	completeSetup(t, h)
	loginGet := doGet(h, "/login", nil)
	loginCsrf, cookies := extractCSRF(t, loginGet)
	loginPost := doPostForm(h, "/login", url.Values{
		"csrf_token": {loginCsrf}, "username": {"sohail"}, "password": {"correct-horse-battery"},
	}, cookies)
	sessionCookies := mergeCookies(cookies, loginPost.Result().Cookies())

	req := httptest.NewRequest(http.MethodGet, "/members?q=nobody", nil)
	req.Header.Set("HX-Request", "true")
	for _, c := range sessionCookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("HX-Request GET /members = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<html") {
		t.Error("HX-Request response includes the full page shell, want just the fragment")
	}
	if !strings.Contains(body, "No members match") {
		t.Error("HTMX search fragment missing no-results text")
	}
}
