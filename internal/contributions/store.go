// Package contributions implements the contribution logging and tracking store
// (docs/SCHEMA.md §3.3, docs/APP_FLOW.md §3). Contributions are soft-deleted by setting deleted_at.
package contributions

import (
	"database/sql"
	"errors"
)

// Contribution represents a recorded payment.
type Contribution struct {
	ID                  int64
	MemberID            int64
	MemberName          string // Joined from members
	AmountMinor         int64
	ContributionMonth   string // "YYYY-MM"
	PaidOn              string // "YYYY-MM-DD"
	RecordedByAdminID   int64
	RecordedByAdminName string // Joined from admins
	DeletedAt           *string
	DeletedByAdminID    *int64
	CreatedAt           string
	UpdatedAt           string
}

// ErrNotFound is returned when a contribution lookup fails.
var ErrNotFound = errors.New("contribution not found")

// Create inserts a new contribution record, ensuring audit tracking.
func Create(conn *sql.DB, memberID int64, amountMinor int64, contributionMonth string, paidOn string, recordedByAdminID int64) (int64, error) {
	if amountMinor <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}
	res, err := conn.Exec(
		`INSERT INTO contributions (member_id, amount_minor, contribution_month, paid_on, recorded_by_admin_id)
		 VALUES (?, ?, ?, ?, ?)`,
		memberID, amountMinor, contributionMonth, paidOn, recordedByAdminID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Get retrieves a contribution by its ID, joining member and admin names.
func Get(conn *sql.DB, id int64) (*Contribution, error) {
	var c Contribution
	var deletedBy NullInt64
	err := conn.QueryRow(
		`SELECT 
			c.id, c.member_id, m.name, c.amount_minor, c.contribution_month, c.paid_on, 
			c.recorded_by_admin_id, a.display_name, c.deleted_at, c.deleted_by_admin_id, c.created_at, c.updated_at
		 FROM contributions c
		 JOIN members m ON c.member_id = m.id
		 JOIN admins a ON c.recorded_by_admin_id = a.id
		 WHERE c.id = ?`,
		id,
	).Scan(
		&c.ID, &c.MemberID, &c.MemberName, &c.AmountMinor, &c.ContributionMonth, &c.PaidOn,
		&c.RecordedByAdminID, &c.RecordedByAdminName, &c.DeletedAt, &deletedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if deletedBy.Valid {
		val := deletedBy.Int64
		c.DeletedByAdminID = &val
	}
	return &c, nil
}

// SoftDelete marks a contribution as deleted, recording the acting admin.
func SoftDelete(conn *sql.DB, id int64, deletedByAdminID int64) error {
	res, err := conn.Exec(
		`UPDATE contributions 
		 SET deleted_at = strftime('%Y-%m-%dT%H:%M:%fZ','now'), 
		     deleted_by_admin_id = ?, 
		     updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') 
		 WHERE id = ? AND deleted_at IS NULL`,
		deletedByAdminID, id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListActiveByMonth returns active contributions for a given month (YYYY-MM).
func ListActiveByMonth(conn *sql.DB, month string) ([]Contribution, error) {
	rows, err := conn.Query(
		`SELECT 
			c.id, c.member_id, m.name, c.amount_minor, c.contribution_month, c.paid_on, 
			c.recorded_by_admin_id, a.display_name, c.created_at, c.updated_at
		 FROM contributions c
		 JOIN members m ON c.member_id = m.id
		 JOIN admins a ON c.recorded_by_admin_id = a.id
		 WHERE c.contribution_month = ? AND c.deleted_at IS NULL
		 ORDER BY c.paid_on DESC, c.id DESC`,
		month,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Contribution
	for rows.Next() {
		var c Contribution
		if err := rows.Scan(
			&c.ID, &c.MemberID, &c.MemberName, &c.AmountMinor, &c.ContributionMonth, &c.PaidOn,
			&c.RecordedByAdminID, &c.RecordedByAdminName, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// ListRecentActive returns the last N active contributions overall.
func ListRecentActive(conn *sql.DB, limit int) ([]Contribution, error) {
	rows, err := conn.Query(
		`SELECT 
			c.id, c.member_id, m.name, c.amount_minor, c.contribution_month, c.paid_on, 
			c.recorded_by_admin_id, a.display_name, c.created_at, c.updated_at
		 FROM contributions c
		 JOIN members m ON c.member_id = m.id
		 JOIN admins a ON c.recorded_by_admin_id = a.id
		 WHERE c.deleted_at IS NULL
		 ORDER BY c.created_at DESC, c.id DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Contribution
	for rows.Next() {
		var c Contribution
		if err := rows.Scan(
			&c.ID, &c.MemberID, &c.MemberName, &c.AmountMinor, &c.ContributionMonth, &c.PaidOn,
			&c.RecordedByAdminID, &c.RecordedByAdminName, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// CheckDuplicate checks if a member already has a contribution logged for a given month.
func CheckDuplicate(conn *sql.DB, memberID int64, month string) (bool, error) {
	var count int
	err := conn.QueryRow(
		`SELECT COUNT(1) FROM contributions WHERE member_id = ? AND contribution_month = ? AND deleted_at IS NULL`,
		memberID, month,
	).Scan(&count)
	return count > 0, err
}

// NullInt64 is a wrapper for nullable integer scans.
type NullInt64 struct {
	Int64 int64
	Valid bool
}

// Scan implements the Scanner interface.
func (ni *NullInt64) Scan(value interface{}) error {
	if value == nil {
		ni.Int64, ni.Valid = 0, false
		return nil
	}
	ni.Valid = true
	switch v := value.(type) {
	case int64:
		ni.Int64 = v
	case int:
		ni.Int64 = int64(v)
	default:
		// Try scanning from sql.NullInt64 or fallback
		var sqlNi sql.NullInt64
		if err := sqlNi.Scan(value); err != nil {
			return err
		}
		ni.Int64 = sqlNi.Int64
		ni.Valid = sqlNi.Valid
	}
	return nil
}
