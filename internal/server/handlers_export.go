package server

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

// handleExportPage renders the export menu.
func (h *authHandlers) handleExportPage(w http.ResponseWriter, r *http.Request) {
	// Check if Litestream backup is configured via R2 bucket env var
	backupActive := os.Getenv("LITESTREAM_R2_BUCKET") != ""
	pages.Export(h.shellData(w, r, "more", "nav.export"), backupActive, "").Render(r.Context(), w)
}

// handleExportDatabase streams a sanitized SQLite database snapshot.
func (h *authHandlers) handleExportDatabase(w http.ResponseWriter, r *http.Request) {
	// Create a temporary file to hold the vacuumed/sanitized database
	tempFile, err := os.CreateTemp("", "sadqa-ledger-export-*.db")
	if err != nil {
		http.Error(w, "failed to create export file", http.StatusInternalServerError)
		return
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// SQLite VACUUM INTO writes a transaction-consistent, clean copy of the DB
	_, err = h.conn.Exec("VACUUM INTO ?", tempPath)
	if err != nil {
		http.Error(w, "failed to generate DB snapshot: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Open the temporary database to redact sensitive columns/tables
	tempDB, err := sql.Open("sqlite", tempPath)
	if err != nil {
		http.Error(w, "failed to sanitize database", http.StatusInternalServerError)
		return
	}
	defer tempDB.Close()

	// 1. Drop sessions table completely (docs/SCHEMA.md §7)
	_, err = tempDB.Exec("DROP TABLE IF EXISTS sessions")
	if err != nil {
		http.Error(w, "failed to drop sessions", http.StatusInternalServerError)
		return
	}

	// 2. Redact/nullify admin password hashes (docs/SCHEMA.md §7)
	_, err = tempDB.Exec("UPDATE admins SET password_hash = ''")
	if err != nil {
		http.Error(w, "failed to redact passwords", http.StatusInternalServerError)
		return
	}

	// Clean up deleted table bytes
	_, _ = tempDB.Exec("VACUUM")
	tempDB.Close()

	// Stream the sanitized database back to the admin
	file, err := os.Open(tempPath)
	if err != nil {
		http.Error(w, "failed to read snapshot", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", `attachment; filename="sadqa-ledger-sanitized.db"`)
	_, _ = io.Copy(w, file)
}

// handleExportContributions streams active contributions as a CSV spreadsheet.
func (h *authHandlers) handleExportContributions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.conn.Query(
		`SELECT c.id, m.name, c.amount_minor, c.contribution_month, c.paid_on, a.display_name, c.created_at
		 FROM contributions c
		 JOIN members m ON c.member_id = m.id
		 JOIN admins a ON c.recorded_by_admin_id = a.id
		 WHERE c.deleted_at IS NULL
		 ORDER BY c.paid_on DESC, c.id DESC`,
	)
	if err != nil {
		http.Error(w, "failed to query contributions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="contributions.csv"`)

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header row
	_ = csvWriter.Write([]string{"ID", "Member Name", "Amount (Major)", "Month", "Paid On", "Recorded By", "Created At"})

	for rows.Next() {
		var id int64
		var memberName, month, paidOn, recordedBy, createdAt string
		var amountMinor int64

		if err := rows.Scan(&id, &memberName, &amountMinor, &month, &paidOn, &recordedBy, &createdAt); err != nil {
			return
		}

		amountMajor := fmt.Sprintf("%.2f", float64(amountMinor)/100.0)
		_ = csvWriter.Write([]string{
			strconv.FormatInt(id, 10),
			memberName,
			amountMajor,
			month,
			paidOn,
			recordedBy,
			createdAt,
		})
	}
}

// handleExportExpenses streams active expenses as a CSV spreadsheet.
func (h *authHandlers) handleExportExpenses(w http.ResponseWriter, r *http.Request) {
	rows, err := h.conn.Query(
		`SELECT e.id, e.description, e.amount_minor, e.expense_date, COALESCE(e.receipt_photo_path, ''), a.display_name, e.created_at
		 FROM expenses e
		 JOIN admins a ON e.recorded_by_admin_id = a.id
		 WHERE e.deleted_at IS NULL
		 ORDER BY e.expense_date DESC, e.id DESC`,
	)
	if err != nil {
		http.Error(w, "failed to query expenses", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="expenses.csv"`)

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header row
	_ = csvWriter.Write([]string{"ID", "Description", "Amount (Major)", "Expense Date", "Receipt Photo Path", "Recorded By", "Created At"})

	for rows.Next() {
		var id int64
		var description, expenseDate, receiptPath, recordedBy, createdAt string
		var amountMinor int64

		if err := rows.Scan(&id, &description, &amountMinor, &expenseDate, &receiptPath, &recordedBy, &createdAt); err != nil {
			return
		}

		amountMajor := fmt.Sprintf("%.2f", float64(amountMinor)/100.0)
		_ = csvWriter.Write([]string{
			strconv.FormatInt(id, 10),
			description,
			amountMajor,
			expenseDate,
			receiptPath,
			recordedBy,
			createdAt,
		})
	}
}
