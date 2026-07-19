package expenses

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
	CREATE TABLE expenses (
		id                    INTEGER PRIMARY KEY,
		description           TEXT NOT NULL,
		amount_minor          INTEGER NOT NULL CHECK (amount_minor > 0),
		expense_date          TEXT NOT NULL,
		receipt_photo_path    TEXT,
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
	return conn
}

func TestCreateAndGetExpense(t *testing.T) {
	conn := openTestDB(t)

	photo := "receipts/abc.png"
	id, err := Create(conn, "Electricity Bill", 45000, "2026-07-15", &photo, 1)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	e, err := Get(conn, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if e.Description != "Electricity Bill" {
		t.Errorf("Description = %q, want %q", e.Description, "Electricity Bill")
	}
	if e.AmountMinor != 45000 {
		t.Errorf("AmountMinor = %d, want 45000", e.AmountMinor)
	}
	if e.ExpenseDate != "2026-07-15" {
		t.Errorf("ExpenseDate = %q, want %q", e.ExpenseDate, "2026-07-15")
	}
	if e.ReceiptPhotoPath == nil || *e.ReceiptPhotoPath != photo {
		t.Errorf("ReceiptPhotoPath = %v, want %q", e.ReceiptPhotoPath, photo)
	}
	if e.RecordedByAdminID != 1 {
		t.Errorf("RecordedByAdminID = %d, want 1", e.RecordedByAdminID)
	}
}

func TestSoftDeleteExpense(t *testing.T) {
	conn := openTestDB(t)

	id, _ := Create(conn, "Electricity", 45000, "2026-07-15", nil, 1)

	if err := SoftDelete(conn, id, 1); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	e, err := Get(conn, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if e.DeletedAt == nil {
		t.Error("DeletedAt is nil after SoftDelete")
	}
	if e.DeletedByAdminID == nil || *e.DeletedByAdminID != 1 {
		t.Errorf("DeletedByAdminID = %v, want 1", e.DeletedByAdminID)
	}

	// Should not be in active list
	list, err := ListActive(conn)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListActive returned %d elements, want 0", len(list))
	}
}
