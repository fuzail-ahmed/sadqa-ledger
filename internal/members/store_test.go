package members

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
	);`
	if _, err := conn.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO admins (id, username, password_hash, display_name) VALUES (1, 'sohail', 'h', 'Sohail')`); err != nil {
		t.Fatalf("insert admin: %v", err)
	}
	return conn
}

func TestCreateDefaultsActiveAndRecordsAdmin(t *testing.T) {
	conn := openTestDB(t)
	id, err := Create(conn, "Farhan", true, 1)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := Get(conn, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !m.IsActive {
		t.Error("new member is not active by default")
	}

	var createdBy, updatedBy int64
	conn.QueryRow(`SELECT created_by_admin_id, updated_by_admin_id FROM members WHERE id = ?`, id).Scan(&createdBy, &updatedBy)
	if createdBy != 1 || updatedBy != 1 {
		t.Errorf("created_by=%d updated_by=%d, want both 1", createdBy, updatedBy)
	}
}

func TestUpdateChangesNameAndStatus(t *testing.T) {
	conn := openTestDB(t)
	id, _ := Create(conn, "Farhan", true, 1)

	if err := Update(conn, id, "Farhan Ahmed", false, 1); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := Get(conn, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Name != "Farhan Ahmed" {
		t.Errorf("Name = %q, want %q", m.Name, "Farhan Ahmed")
	}
	if m.IsActive {
		t.Error("member still active after Update(isActive=false)")
	}
}

func TestUpdateUnknownIDReturnsNotFound(t *testing.T) {
	conn := openTestDB(t)
	if err := Update(conn, 999, "Nobody", true, 1); err != ErrNotFound {
		t.Errorf("Update(unknown id) = %v, want ErrNotFound", err)
	}
}

func TestSetActiveSoftDeleteAndReactivate(t *testing.T) {
	conn := openTestDB(t)
	id, _ := Create(conn, "Farhan", true, 1)

	if err := SetActive(conn, id, false, 1); err != nil {
		t.Fatalf("SetActive(false): %v", err)
	}
	m, _ := Get(conn, id)
	if m.IsActive {
		t.Fatal("member still active after deactivation")
	}

	// Soft delete: the row itself must survive (docs/SCHEMA.md §3.2 — history retained).
	var count int
	conn.QueryRow(`SELECT COUNT(*) FROM members WHERE id = ?`, id).Scan(&count)
	if count != 1 {
		t.Fatal("member row was removed instead of soft-deleted")
	}

	if err := SetActive(conn, id, true, 1); err != nil {
		t.Fatalf("SetActive(true): %v", err)
	}
	m, _ = Get(conn, id)
	if !m.IsActive {
		t.Error("member still inactive after reactivation")
	}
}

func TestSetActiveUnknownIDReturnsNotFound(t *testing.T) {
	conn := openTestDB(t)
	if err := SetActive(conn, 999, false, 1); err != ErrNotFound {
		t.Errorf("SetActive(unknown id) = %v, want ErrNotFound", err)
	}
}

func TestListFiltersByNameCaseInsensitive(t *testing.T) {
	conn := openTestDB(t)
	Create(conn, "Abdul Rahman", true, 1)
	Create(conn, "Farhan", true, 1)

	all, err := List(conn, "")
	if err != nil {
		t.Fatalf("List(\"\"): %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("List(\"\") returned %d members, want 2", len(all))
	}

	match, err := List(conn, "abdul")
	if err != nil {
		t.Fatalf("List(abdul): %v", err)
	}
	if len(match) != 1 || match[0].Name != "Abdul Rahman" {
		t.Errorf("List(abdul) = %+v, want just Abdul Rahman", match)
	}

	noMatch, err := List(conn, "zzz")
	if err != nil {
		t.Fatalf("List(zzz): %v", err)
	}
	if len(noMatch) != 0 {
		t.Errorf("List(zzz) = %+v, want empty", noMatch)
	}
}

func TestListIncludesInactiveMembers(t *testing.T) {
	conn := openTestDB(t)
	id, _ := Create(conn, "Farhan", true, 1)
	SetActive(conn, id, false, 1)

	list, err := List(conn, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].IsActive {
		t.Errorf("List() = %+v, want one inactive member still present", list)
	}
}

func TestCreateWithInactiveStatus(t *testing.T) {
	conn := openTestDB(t)
	id, err := Create(conn, "Tariq", false, 1)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := Get(conn, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.IsActive {
		t.Error("new member created with isActive=false is active, want inactive")
	}
}
