// Package db manages the SQLite connection and the migration runner. DSN
// pragma syntax is driver-specific — see docs/SCHEMA.md §1a: modernc.org/sqlite
// uses repeatable ?_pragma=name(value) params, not mattn's ?_journal_mode=WAL
// style, which modernc silently ignores rather than erroring on.
package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open connects to the SQLite database at path with WAL mode, foreign key
// enforcement, and a busy timeout set at the driver level (docs/SCHEMA.md §1a).
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// ponytail: single writer keeps SQLite lock handling trivial for 2-3
	// admins; raise if read concurrency ever profiles as a bottleneck.
	conn.SetMaxOpenConns(1)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return conn, nil
}
