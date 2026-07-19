package auth

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	sqldb "github.com/fuzail-ahmed/sadqa-ledger/internal/db"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sqldb.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	schema := `
	CREATE TABLE admins (
		id INTEGER PRIMARY KEY, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL,
		display_name TEXT NOT NULL, language_pref TEXT NOT NULL DEFAULT 'en',
		is_active INTEGER NOT NULL DEFAULT 1
	);
	CREATE TABLE group_settings (
		id INTEGER PRIMARY KEY CHECK (id = 1), group_name TEXT NOT NULL,
		currency_code TEXT NOT NULL DEFAULT 'INR', currency_symbol TEXT NOT NULL DEFAULT '?',
		public_token TEXT NOT NULL UNIQUE
	);
	CREATE TABLE sessions (
		token_hash TEXT PRIMARY KEY, admin_id INTEGER NOT NULL REFERENCES admins(id),
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		expires_at TEXT NOT NULL,
		last_seen_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		user_agent TEXT, ip_address TEXT
	);`
	if _, err := conn.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return conn
}

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if strings.Contains(hash, "correct-horse-battery-staple") {
		t.Fatal("hash contains the plaintext password")
	}
	if !VerifyPassword(hash, "correct-horse-battery-staple") {
		t.Error("VerifyPassword: correct password rejected")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Error("VerifyPassword: wrong password accepted")
	}
}

func TestCreateSessionStoresHashNotRawToken(t *testing.T) {
	conn := openTestDB(t)
	adminID, err := CreateAdmin(conn, "sohail", "irrelevant-hash", "Sohail")
	if err != nil {
		t.Fatalf("CreateAdmin: %v", err)
	}

	rawToken, err := CreateSession(conn, adminID, time.Hour, "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	var storedHash string
	if err := conn.QueryRow(`SELECT token_hash FROM sessions WHERE admin_id = ?`, adminID).Scan(&storedHash); err != nil {
		t.Fatalf("query session: %v", err)
	}
	if storedHash == rawToken {
		t.Fatal("raw token stored directly in sessions.token_hash")
	}
	if storedHash != hashToken(rawToken) {
		t.Error("stored value is not SHA-256(rawToken)")
	}

	// Sanity: the raw token must not appear anywhere in the sessions table.
	var rawTokenAppearsInDB int
	conn.QueryRow(`SELECT COUNT(*) FROM sessions WHERE token_hash = ?`, rawToken).Scan(&rawTokenAppearsInDB)
	if rawTokenAppearsInDB != 0 {
		t.Error("raw token found stored as a token_hash value")
	}
}

func TestLookupSessionSuccessAndFailure(t *testing.T) {
	conn := openTestDB(t)
	adminID, _ := CreateAdmin(conn, "sohail", "hash", "Sohail")
	rawToken, err := CreateSession(conn, adminID, time.Hour, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	sess, err := LookupSession(conn, rawToken)
	if err != nil {
		t.Fatalf("LookupSession(valid): %v", err)
	}
	if sess.AdminID != adminID {
		t.Errorf("AdminID = %d, want %d", sess.AdminID, adminID)
	}

	if _, err := LookupSession(conn, "not-a-real-token"); err != ErrNoSession {
		t.Errorf("LookupSession(garbage) = %v, want ErrNoSession", err)
	}
}

func TestLookupSessionRejectsExpired(t *testing.T) {
	conn := openTestDB(t)
	adminID, _ := CreateAdmin(conn, "sohail", "hash", "Sohail")
	// Negative lifetime: already expired the instant it's created.
	rawToken, err := CreateSession(conn, adminID, -time.Hour, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if _, err := LookupSession(conn, rawToken); err != ErrNoSession {
		t.Errorf("LookupSession(expired) = %v, want ErrNoSession", err)
	}

	var count int
	conn.QueryRow(`SELECT COUNT(*) FROM sessions WHERE token_hash = ?`, hashToken(rawToken)).Scan(&count)
	if count != 0 {
		t.Error("expired session row was not cleaned up on lookup")
	}
}

func TestDeleteSessionLogsOutServerSide(t *testing.T) {
	conn := openTestDB(t)
	adminID, _ := CreateAdmin(conn, "sohail", "hash", "Sohail")
	rawToken, _ := CreateSession(conn, adminID, time.Hour, "", "")

	if err := DeleteSession(conn, rawToken); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if _, err := LookupSession(conn, rawToken); err != ErrNoSession {
		t.Errorf("session still valid after DeleteSession: %v", err)
	}
}

func TestDeleteExpiredSessions(t *testing.T) {
	conn := openTestDB(t)
	adminID, _ := CreateAdmin(conn, "sohail", "hash", "Sohail")
	CreateSession(conn, adminID, -time.Hour, "", "") // expired
	CreateSession(conn, adminID, time.Hour, "", "")  // still valid

	n, err := DeleteExpiredSessions(conn)
	if err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}
	if n != 1 {
		t.Errorf("deleted %d rows, want 1", n)
	}

	var remaining int
	conn.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&remaining)
	if remaining != 1 {
		t.Errorf("%d sessions remain, want 1", remaining)
	}
}

func TestGroupSettingsExist(t *testing.T) {
	conn := openTestDB(t)
	exists, err := GroupSettingsExist(conn)
	if err != nil {
		t.Fatalf("GroupSettingsExist: %v", err)
	}
	if exists {
		t.Fatal("GroupSettingsExist = true on empty DB")
	}

	if _, err := CreateGroupInSetup(conn, "admin", "hash", "Admin", "Test Masjid", "INR", "₹"); err != nil {
		t.Fatalf("CreateGroupInSetup: %v", err)
	}

	exists, err = GroupSettingsExist(conn)
	if err != nil {
		t.Fatalf("GroupSettingsExist: %v", err)
	}
	if !exists {
		t.Fatal("GroupSettingsExist = false after CreateGroupInSetup")
	}
}

func TestFindActiveAdminByUsernameSkipsDeactivated(t *testing.T) {
	conn := openTestDB(t)
	if _, err := conn.Exec(`INSERT INTO admins (username, password_hash, display_name, is_active) VALUES ('inactive', 'h', 'Old Admin', 0)`); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if _, err := FindActiveAdminByUsername(conn, "inactive"); err != ErrAdminNotFound {
		t.Errorf("FindActiveAdminByUsername(deactivated) = %v, want ErrAdminNotFound", err)
	}
	if _, err := FindActiveAdminByUsername(conn, "nonexistent"); err != ErrAdminNotFound {
		t.Errorf("FindActiveAdminByUsername(unknown) = %v, want ErrAdminNotFound", err)
	}
}
