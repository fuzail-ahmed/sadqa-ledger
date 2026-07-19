package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
)

var migrationFileRe = regexp.MustCompile(`^(\d+)_.*\.sql$`)

// Migrate applies every pending migration in migrations (sorted by numeric
// filename prefix) that isn't already recorded in schema_migrations, per
// docs/SCHEMA.md §8. Safe to call repeatedly: already-applied migrations are
// skipped.
func Migrate(conn *sql.DB, migrations fs.FS) error {
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations, ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type pendingMigration struct {
		version int
		name    string
	}
	var files []pendingMigration
	for _, e := range entries {
		m := migrationFileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		version, err := strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("parse version from %s: %w", e.Name(), err)
		}
		files = append(files, pendingMigration{version, e.Name()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })

	for _, f := range files {
		var applied int
		err := conn.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, f.version).Scan(&applied)
		if err == nil {
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %d: %w", f.version, err)
		}

		content, err := fs.ReadFile(migrations, f.name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f.name, err)
		}

		tx, err := conn.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", f.version, err)
		}
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %d (%s): %w", f.version, f.name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, f.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", f.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", f.version, err)
		}
	}

	return nil
}
