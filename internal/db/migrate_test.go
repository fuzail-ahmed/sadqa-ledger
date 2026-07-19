package db

import (
	"database/sql"
	"testing"
	"testing/fstest"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestMigrateAppliesInOrder(t *testing.T) {
	conn := openTestDB(t)

	fsys := fstest.MapFS{
		"0002_second.sql": {Data: []byte(`CREATE TABLE second (id INTEGER PRIMARY KEY);`)},
		"0001_first.sql":  {Data: []byte(`CREATE TABLE first (id INTEGER PRIMARY KEY);`)},
	}

	if err := Migrate(conn, fsys); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	for _, table := range []string{"first", "second"} {
		var name string
		if err := conn.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name); err != nil {
			t.Errorf("table %s not created: %v", table, err)
		}
	}

	var versions []int
	rows, err := conn.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan version: %v", err)
		}
		versions = append(versions, v)
	}
	if len(versions) != 2 || versions[0] != 1 || versions[1] != 2 {
		t.Fatalf("expected versions [1 2], got %v", versions)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	conn := openTestDB(t)

	fsys := fstest.MapFS{
		"0001_first.sql": {Data: []byte(`CREATE TABLE first (id INTEGER PRIMARY KEY);`)},
	}

	if err := Migrate(conn, fsys); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	// Second run must not try to re-run 0001_first.sql (which would fail:
	// "table first already exists") — this is the idempotency guarantee.
	if err := Migrate(conn, fsys); err != nil {
		t.Fatalf("second Migrate (should be a no-op): %v", err)
	}

	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 recorded migration, got %d", count)
	}
}
