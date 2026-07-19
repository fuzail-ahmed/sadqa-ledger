// Package dashboard retrieves the status and statistics for the main dashboard view
// (docs/SCHEMA.md, docs/APP_FLOW.md §2).
package dashboard

import (
	"database/sql"
	"fmt"
)

// ChecklistItem represents a member's payment status for the month.
type ChecklistItem struct {
	MemberID    int64
	MemberName  string
	Paid        bool
	AmountMinor int64
	PaidOn      string // Date of payment (latest if multiple)
}

// ActivityItem represents a contribution or expense event.
type ActivityItem struct {
	Type        string // "contribution" | "expense"
	ID          int64
	AmountMinor int64
	CreatedAt   string
	Detail      string // Member name for contributions, description for expenses
	RecordedBy  string // Admin display name
}

// DashboardData holds all the data needed to render the home dashboard.
type DashboardData struct {
	ThisMonthCollected int64
	AllTimeCollected   int64
	AllTimeExpenses    int64
	CurrentBalance     int64
	Checklist          []ChecklistItem
	RecentActivity     []ActivityItem
}

// GetDashboardData aggregates statistics, builds the checklist, and retrieves recent activity.
func GetDashboardData(conn *sql.DB, month string) (*DashboardData, error) {
	data := &DashboardData{}

	// 1. This Month's Collection
	err := conn.QueryRow(
		`SELECT COALESCE(SUM(amount_minor), 0) FROM contributions WHERE contribution_month = ? AND deleted_at IS NULL`,
		month,
	).Scan(&data.ThisMonthCollected)
	if err != nil {
		return nil, fmt.Errorf("this month collected: %w", err)
	}

	// 2. All-Time Total Collected
	err = conn.QueryRow(
		`SELECT COALESCE(SUM(amount_minor), 0) FROM contributions WHERE deleted_at IS NULL`,
	).Scan(&data.AllTimeCollected)
	if err != nil {
		return nil, fmt.Errorf("all-time collected: %w", err)
	}

	// 3. All-Time Total Expenses
	err = conn.QueryRow(
		`SELECT COALESCE(SUM(amount_minor), 0) FROM expenses WHERE deleted_at IS NULL`,
	).Scan(&data.AllTimeExpenses)
	if err != nil {
		return nil, fmt.Errorf("all-time expenses: %w", err)
	}

	// 4. Current Balance
	data.CurrentBalance = data.AllTimeCollected - data.AllTimeExpenses

	// 5. Checklist: active members and their payments for the month (summed if multiple)
	rows, err := conn.Query(
		`SELECT 
			m.id, m.name, 
			CASE WHEN c.id IS NOT NULL THEN 1 ELSE 0 END as paid,
			COALESCE(SUM(c.amount_minor), 0) as amount,
			COALESCE(MAX(c.paid_on), '') as paid_on
		 FROM members m
		 LEFT JOIN contributions c ON m.id = c.member_id AND c.contribution_month = ? AND c.deleted_at IS NULL
		 WHERE m.is_active = 1
		 GROUP BY m.id, m.name
		 ORDER BY m.name COLLATE NOCASE`,
		month,
	)
	if err != nil {
		return nil, fmt.Errorf("checklist query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item ChecklistItem
		var paid int
		if err := rows.Scan(&item.MemberID, &item.MemberName, &paid, &item.AmountMinor, &item.PaidOn); err != nil {
			return nil, fmt.Errorf("checklist scan: %w", err)
		}
		item.Paid = paid == 1
		data.Checklist = append(data.Checklist, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("checklist rows error: %w", err)
	}

	// 6. Recent Activity Feed
	activityRows, err := conn.Query(
		`SELECT type, id, amount_minor, created_at, detail, recorded_by FROM (
			SELECT 'contribution' as type, c.id as id, c.amount_minor, c.created_at, m.name as detail, a.display_name as recorded_by 
			FROM contributions c 
			JOIN members m ON c.member_id = m.id 
			JOIN admins a ON c.recorded_by_admin_id = a.id 
			WHERE c.deleted_at IS NULL
			UNION ALL
			SELECT 'expense' as type, e.id as id, e.amount_minor, e.created_at, e.description as detail, a.display_name as recorded_by 
			FROM expenses e 
			JOIN admins a ON e.recorded_by_admin_id = a.id 
			WHERE e.deleted_at IS NULL
		) ORDER BY created_at DESC, id DESC
		LIMIT 10`,
	)
	if err != nil {
		return nil, fmt.Errorf("activity query: %w", err)
	}
	defer activityRows.Close()

	for activityRows.Next() {
		var item ActivityItem
		if err := activityRows.Scan(&item.Type, &item.ID, &item.AmountMinor, &item.CreatedAt, &item.Detail, &item.RecordedBy); err != nil {
			return nil, fmt.Errorf("activity scan: %w", err)
		}
		data.RecentActivity = append(data.RecentActivity, item)
	}
	if err := activityRows.Err(); err != nil {
		return nil, fmt.Errorf("activity rows error: %w", err)
	}

	return data, nil
}
