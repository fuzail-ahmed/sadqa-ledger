// Package expenses implements the expense logging and receipt storage
// (docs/SCHEMA.md §3.4, docs/APP_FLOW.md §5). Expenses are soft-deleted.
package expenses

import (
	"database/sql"
	"errors"
)

// Expense represents a recorded community expense.
type Expense struct {
	ID                  int64
	Description         string
	AmountMinor         int64
	ExpenseDate         string // "YYYY-MM-DD"
	ReceiptPhotoPath    *string
	RecordedByAdminID   int64
	RecordedByAdminName string // Joined from admins
	DeletedAt           *string
	DeletedByAdminID    *int64
	CreatedAt           string
	UpdatedAt           string
}

var ErrNotFound = errors.New("expense not found")

// Create inserts a new expense record.
func Create(
	conn *sql.DB,
	description string,
	amountMinor int64,
	expenseDate string,
	receiptPhotoPath *string,
	recordedByAdminID int64,
) (int64, error) {
	if amountMinor <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}
	if description == "" {
		return 0, errors.New("description is required")
	}

	var photoPath sql.NullString
	if receiptPhotoPath != nil {
		photoPath = sql.NullString{String: *receiptPhotoPath, Valid: true}
	}

	res, err := conn.Exec(
		`INSERT INTO expenses (description, amount_minor, expense_date, receipt_photo_path, recorded_by_admin_id)
		 VALUES (?, ?, ?, ?, ?)`,
		description, amountMinor, expenseDate, photoPath, recordedByAdminID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Get retrieves a single expense by ID, joining admin name.
func Get(conn *sql.DB, id int64) (*Expense, error) {
	var e Expense
	var photoPath sql.NullString
	var deletedBy NullInt64

	err := conn.QueryRow(
		`SELECT 
			e.id, e.description, e.amount_minor, e.expense_date, e.receipt_photo_path, 
			e.recorded_by_admin_id, a.display_name, e.deleted_at, e.deleted_by_admin_id, e.created_at, e.updated_at
		 FROM expenses e
		 JOIN admins a ON e.recorded_by_admin_id = a.id
		 WHERE e.id = ?`,
		id,
	).Scan(
		&e.ID, &e.Description, &e.AmountMinor, &e.ExpenseDate, &photoPath,
		&e.RecordedByAdminID, &e.RecordedByAdminName, &e.DeletedAt, &deletedBy, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if photoPath.Valid {
		e.ReceiptPhotoPath = &photoPath.String
	}
	if deletedBy.Valid {
		val := deletedBy.Int64
		e.DeletedByAdminID = &val
	}

	return &e, nil
}

// SoftDelete marks an expense as deleted, recording the acting admin.
func SoftDelete(conn *sql.DB, id int64, deletedByAdminID int64) error {
	res, err := conn.Exec(
		`UPDATE expenses 
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

// ListActive returns all active (non-deleted) expenses, ordered newest first.
func ListActive(conn *sql.DB) ([]Expense, error) {
	rows, err := conn.Query(
		`SELECT 
			e.id, e.description, e.amount_minor, e.expense_date, e.receipt_photo_path, 
			e.recorded_by_admin_id, a.display_name, e.created_at, e.updated_at
		 FROM expenses e
		 JOIN admins a ON e.recorded_by_admin_id = a.id
		 WHERE e.deleted_at IS NULL
		 ORDER BY e.expense_date DESC, e.id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Expense
	for rows.Next() {
		var e Expense
		var photoPath sql.NullString
		if err := rows.Scan(
			&e.ID, &e.Description, &e.AmountMinor, &e.ExpenseDate, &photoPath,
			&e.RecordedByAdminID, &e.RecordedByAdminName, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if photoPath.Valid {
			e.ReceiptPhotoPath = &photoPath.String
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

// NullInt64 is a wrapper for scan.
type NullInt64 struct {
	Int64 int64
	Valid bool
}

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
		var sqlNi sql.NullInt64
		if err := sqlNi.Scan(value); err != nil {
			return err
		}
		ni.Int64 = sqlNi.Int64
		ni.Valid = sqlNi.Valid
	}
	return nil
}
