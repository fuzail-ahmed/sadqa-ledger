package dashboard

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
		notes TEXT,
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

func TestGetDashboardData(t *testing.T) {
	conn := openTestDB(t)

	// Insert members
	_, _ = conn.Exec(`INSERT INTO members (id, name, is_active) VALUES (1, 'Farhan', 1), (2, 'Abdul', 1), (3, 'Inactive User', 0)`)

	// Insert contributions
	_, _ = conn.Exec(`INSERT INTO contributions (id, member_id, amount_minor, contribution_month, paid_on, recorded_by_admin_id) 
	                 VALUES (1, 1, 50000, '2026-07', '2026-07-15', 1),
	                        (2, 1, 20000, '2026-07', '2026-07-16', 1), -- multiple payments for checklist sum
	                        (3, 2, 30000, '2026-06', '2026-06-15', 1)`) // diff month

	// Insert expenses
	_, _ = conn.Exec(`INSERT INTO expenses (id, description, amount_minor, expense_date, recorded_by_admin_id) 
	                 VALUES (1, 'Electricity', 15000, '2026-07-10', 1)`)

	data, err := GetDashboardData(conn, "2026-07")
	if err != nil {
		t.Fatalf("GetDashboardData: %v", err)
	}

	// Verify stats
	if data.ThisMonthCollected != 70000 {
		t.Errorf("ThisMonthCollected = %d, want 70000", data.ThisMonthCollected)
	}
	if data.AllTimeCollected != 100000 {
		t.Errorf("AllTimeCollected = %d, want 100000", data.AllTimeCollected)
	}
	if data.AllTimeExpenses != 15000 {
		t.Errorf("AllTimeExpenses = %d, want 15000", data.AllTimeExpenses)
	}
	if data.CurrentBalance != 85000 {
		t.Errorf("CurrentBalance = %d, want 85000", data.CurrentBalance)
	}

	// Verify checklist (inactive member must be excluded)
	if len(data.Checklist) != 2 {
		t.Fatalf("Checklist length = %d, want 2 active members", len(data.Checklist))
	}

	// Checklist ordering collation collation check (Abdul then Farhan)
	if data.Checklist[0].MemberName != "Abdul" || data.Checklist[1].MemberName != "Farhan" {
		t.Errorf("Checklist ordering incorrect: [0]=%s, [1]=%s", data.Checklist[0].MemberName, data.Checklist[1].MemberName)
	}

	// Farhan checklist validation
	farhanItem := data.Checklist[1]
	if !farhanItem.Paid {
		t.Error("Farhan.Paid is false, want true")
	}
	if farhanItem.AmountMinor != 70000 {
		t.Errorf("Farhan.AmountMinor = %d, want 70000 (50000 + 20000)", farhanItem.AmountMinor)
	}

	// Abdul checklist validation
	abdulItem := data.Checklist[0]
	if abdulItem.Paid {
		t.Error("Abdul.Paid is true for July, want false")
	}

	// Verify activity
	if len(data.RecentActivity) != 4 {
		t.Errorf("RecentActivity length = %d, want 4 items (3 contributions + 1 expense)", len(data.RecentActivity))
	}
}
