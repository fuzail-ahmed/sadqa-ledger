package contributions

import (
	"database/sql"
	"testing"

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
	CREATE TABLE members (
		id INTEGER PRIMARY KEY, name TEXT NOT NULL, is_active INTEGER NOT NULL DEFAULT 1,
		notes TEXT, created_by_admin_id INTEGER REFERENCES admins(id),
		updated_by_admin_id INTEGER REFERENCES admins(id),
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	);
	CREATE TABLE contributions (
		id                    INTEGER PRIMARY KEY,
		member_id             INTEGER NOT NULL REFERENCES members(id),
		amount_minor          INTEGER NOT NULL CHECK (amount_minor > 0),
		contribution_month    TEXT NOT NULL,
		paid_on               TEXT NOT NULL,
		recorded_by_admin_id  INTEGER NOT NULL REFERENCES admins(id),
		deleted_at            TEXT,
		deleted_by_admin_id   INTEGER REFERENCES admins(id),
		created_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	);`
	if _, err := conn.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO admins (id, username, password_hash, display_name) VALUES (1, 'sohail', 'h', 'Sohail')`); err != nil {
		t.Fatalf("insert admin: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO members (id, name, is_active) VALUES (1, 'Farhan', 1)`); err != nil {
		t.Fatalf("insert member: %v", err)
	}
	return conn
}

func TestCreateContribution(t *testing.T) {
	conn := openTestDB(t)

	id, err := Create(conn, 1, 50000, "2026-07", "2026-07-15", 1)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c, err := Get(conn, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if c.MemberID != 1 {
		t.Errorf("MemberID = %d, want 1", c.MemberID)
	}
	if c.MemberName != "Farhan" {
		t.Errorf("MemberName = %q, want %q", c.MemberName, "Farhan")
	}
	if c.AmountMinor != 50000 {
		t.Errorf("AmountMinor = %d, want 50000", c.AmountMinor)
	}
	if c.ContributionMonth != "2026-07" {
		t.Errorf("ContributionMonth = %q, want %q", c.ContributionMonth, "2026-07")
	}
	if c.PaidOn != "2026-07-15" {
		t.Errorf("PaidOn = %q, want %q", c.PaidOn, "2026-07-15")
	}
	if c.RecordedByAdminID != 1 {
		t.Errorf("RecordedByAdminID = %d, want 1", c.RecordedByAdminID)
	}
}

func TestSoftDeleteContribution(t *testing.T) {
	conn := openTestDB(t)

	id, _ := Create(conn, 1, 50000, "2026-07", "2026-07-15", 1)

	if err := SoftDelete(conn, id, 1); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	c, err := Get(conn, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if c.DeletedAt == nil {
		t.Error("DeletedAt is nil after SoftDelete")
	}
	if c.DeletedByAdminID == nil || *c.DeletedByAdminID != 1 {
		t.Errorf("DeletedByAdminID = %v, want 1", c.DeletedByAdminID)
	}

	// Should not be in active lists
	list, err := ListActiveByMonth(conn, "2026-07")
	if err != nil {
		t.Fatalf("ListActiveByMonth: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListActiveByMonth returned %d rows, want 0", len(list))
	}
}

func TestCheckDuplicate(t *testing.T) {
	conn := openTestDB(t)

	has, err := CheckDuplicate(conn, 1, "2026-07")
	if err != nil {
		t.Fatalf("CheckDuplicate: %v", err)
	}
	if has {
		t.Error("CheckDuplicate returned true for empty DB")
	}

	_, _ = Create(conn, 1, 50000, "2026-07", "2026-07-15", 1)

	has, err = CheckDuplicate(conn, 1, "2026-07")
	if err != nil {
		t.Fatalf("CheckDuplicate: %v", err)
	}
	if !has {
		t.Error("CheckDuplicate returned false after insertion")
	}
}
