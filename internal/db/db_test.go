package db

import "testing"

// TestOpenSetsPragmas verifies the pragmas actually took effect, since
// modernc.org/sqlite silently ignores mattn-style DSN params instead of
// erroring (docs/SCHEMA.md §1a) — a wrong DSN would pass Open() but leave
// WAL/foreign keys off, so this has to be checked at runtime, not assumed.
func TestOpenSetsPragmas(t *testing.T) {
	conn := openTestDB(t)

	var journalMode string
	if err := conn.QueryRow(`PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}

	var foreignKeys int
	if err := conn.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("foreign_keys = %d, want 1", foreignKeys)
	}
}

func TestForeignKeysEnforced(t *testing.T) {
	conn := openTestDB(t)

	if _, err := conn.Exec(`CREATE TABLE parent (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if _, err := conn.Exec(`CREATE TABLE child (id INTEGER PRIMARY KEY, parent_id INTEGER REFERENCES parent(id))`); err != nil {
		t.Fatalf("create child: %v", err)
	}

	_, err := conn.Exec(`INSERT INTO child (id, parent_id) VALUES (1, 999)`)
	if err == nil {
		t.Fatal("expected foreign key violation, got no error")
	}
}
