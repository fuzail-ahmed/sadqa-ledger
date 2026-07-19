package server

import (
	"database/sql"
	"encoding/csv"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestExportFlow(t *testing.T) {
	h, _ := newTestServer(t)
	completeSetup(t, h)
	loginGet := doGet(h, "/login", nil)
	loginCsrf, cookies := extractCSRF(t, loginGet)
	loginPost := doPostForm(h, "/login", url.Values{
		"csrf_token": {loginCsrf}, "username": {"sohail"}, "password": {"correct-horse-battery"},
	}, cookies)
	sessionCookies := mergeCookies(cookies, loginPost.Result().Cookies())

	// --- 1. Get export page ---
	exportGet := doGet(h, "/export", sessionCookies)
	if exportGet.Code != 200 {
		t.Fatalf("GET /export = %d, want 200", exportGet.Code)
	}
	if !strings.Contains(exportGet.Body.String(), "Export your group's data") {
		t.Error("Export page missing intro text")
	}

	// --- 2. Download sanitized database snapshot ---
	dbGet := doGet(h, "/export/database", sessionCookies)
	if dbGet.Code != 200 {
		t.Fatalf("GET /export/database = %d, want 200", dbGet.Code)
	}

	// Save downloaded bytes to temp file to query it
	tempFile, err := os.CreateTemp("", "sadqa-test-download-*.db")
	if err != nil {
		t.Fatalf("Temp file creation: %v", err)
	}
	tempPath := tempFile.Name()
	_, _ = tempFile.Write(dbGet.Body.Bytes())
	tempFile.Close()
	defer os.Remove(tempPath)

	// Verify database contents
	testDB, err := sql.Open("sqlite", tempPath)
	if err != nil {
		t.Fatalf("sql.Open temp DB: %v", err)
	}
	defer testDB.Close()

	// Verify sessions table is dropped
	var sessionTableExists int
	_ = testDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&sessionTableExists)
	if sessionTableExists != 0 {
		t.Error("Redaction failed: sessions table still exists in exported database snapshot")
	}

	// Verify passwords are redacted
	var passwordHash string
	err = testDB.QueryRow("SELECT password_hash FROM admins WHERE username='sohail'").Scan(&passwordHash)
	if err != nil {
		t.Fatalf("Query exported admins: %v", err)
	}
	if passwordHash != "" {
		t.Errorf("Redaction failed: password hash in export is %q, want empty string", passwordHash)
	}
	testDB.Close()

	// --- 3. Download contributions CSV ---
	contribsGet := doGet(h, "/export/contributions", sessionCookies)
	if contribsGet.Code != 200 {
		t.Fatalf("GET /export/contributions = %d, want 200", contribsGet.Code)
	}
	cReader := csv.NewReader(strings.NewReader(contribsGet.Body.String()))
	cRecords, err := cReader.ReadAll()
	if err != nil {
		t.Fatalf("Read CSV: %v", err)
	}
	if len(cRecords) == 0 || cRecords[0][1] != "Member Name" {
		t.Errorf("Invalid CSV headers: %v", cRecords[0])
	}

	// --- 4. Download expenses CSV ---
	expensesGet := doGet(h, "/export/expenses", sessionCookies)
	if expensesGet.Code != 200 {
		t.Fatalf("GET /export/expenses = %d, want 200", expensesGet.Code)
	}
	eReader := csv.NewReader(strings.NewReader(expensesGet.Body.String()))
	eRecords, err := eReader.ReadAll()
	if err != nil {
		t.Fatalf("Read CSV: %v", err)
	}
	if len(eRecords) == 0 || eRecords[0][1] != "Description" {
		t.Errorf("Invalid CSV headers: %v", eRecords[0])
	}
}
