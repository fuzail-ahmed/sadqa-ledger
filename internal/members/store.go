// Package members implements the member roster: create, edit, and
// deactivate/reactivate (docs/SCHEMA.md §3.2, docs/APP_FLOW.md §4). Members
// are never hard-deleted — is_active is the soft-delete flag, per
// docs/SCHEMA.md's note that history must survive a member leaving.
package members

import (
	"database/sql"
	"errors"
)

// Member is a row from the members table.
type Member struct {
	ID       int64
	Name     string
	IsActive bool
}

// ErrNotFound is returned when a member id doesn't match any row.
var ErrNotFound = errors.New("member not found")

// List returns members ordered by name, optionally filtered to those whose
// name contains query (case-insensitive substring match, docs/APP_FLOW.md
// §4's live search). An empty query returns every member, active or not —
// the list page shows both with a status badge.
func List(conn *sql.DB, query string) ([]Member, error) {
	rows, err := conn.Query(
		`SELECT id, name, is_active FROM members WHERE name LIKE '%' || ? || '%' COLLATE NOCASE ORDER BY name`,
		query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Member
	for rows.Next() {
		var m Member
		var active int
		if err := rows.Scan(&m.ID, &m.Name, &active); err != nil {
			return nil, err
		}
		m.IsActive = active == 1
		list = append(list, m)
	}
	return list, rows.Err()
}

// Get looks up a single member by id.
func Get(conn *sql.DB, id int64) (*Member, error) {
	var m Member
	var active int
	err := conn.QueryRow(`SELECT id, name, is_active FROM members WHERE id = ?`, id).Scan(&m.ID, &m.Name, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	m.IsActive = active == 1
	return &m, nil
}

// Create inserts a new member with the specified name and active status,
// recording the acting admin.
func Create(conn *sql.DB, name string, isActive bool, createdByAdminID int64) (int64, error) {
	res, err := conn.Exec(
		`INSERT INTO members (name, is_active, created_by_admin_id, updated_by_admin_id) VALUES (?, ?, ?, ?)`,
		name, isActive, createdByAdminID, createdByAdminID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Update changes a member's name and active status, recording the acting
// admin and bumping updated_at.
func Update(conn *sql.DB, id int64, name string, isActive bool, updatedByAdminID int64) error {
	res, err := conn.Exec(
		`UPDATE members SET name = ?, is_active = ?, updated_by_admin_id = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id = ?`,
		name, isActive, updatedByAdminID, id,
	)
	if err != nil {
		return err
	}
	return checkRowAffected(res)
}

// SetActive toggles a member's active flag — the Deactivate/Reactivate quick
// action on the list row (docs/APP_FLOW.md §4). Deactivating does not touch
// any other table: historical contributions reference the member id
// regardless of is_active (docs/SCHEMA.md §3.2/§3.3).
func SetActive(conn *sql.DB, id int64, isActive bool, updatedByAdminID int64) error {
	res, err := conn.Exec(
		`UPDATE members SET is_active = ?, updated_by_admin_id = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id = ?`,
		isActive, updatedByAdminID, id,
	)
	if err != nil {
		return err
	}
	return checkRowAffected(res)
}

func checkRowAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
