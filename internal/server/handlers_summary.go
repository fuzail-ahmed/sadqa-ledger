package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/settings"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

func formatMoney(amountMinor int64, symbol string) string {
	if amountMinor%100 == 0 {
		return fmt.Sprintf("%s%d", symbol, amountMinor/100)
	}
	return fmt.Sprintf("%s%.2f", symbol, float64(amountMinor)/100.0)
}

// handleSummaryPage renders the main monthly summary screen.
func (h *authHandlers) handleSummaryPage(w http.ResponseWriter, r *http.Request) {
	currentMonth := time.Now().Format("2006-01")
	text, err := h.buildSummaryText(currentMonth)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pages.Summary(h.shellData(w, r, "more", "nav.summary"), currentMonth, text).Render(r.Context(), w)
}

// handleSummaryGenerate serves HTMX requests to generate summary text for a chosen month.
func (h *authHandlers) handleSummaryGenerate(w http.ResponseWriter, r *http.Request) {
	month := strings.TrimSpace(r.URL.Query().Get("summary_month"))
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	text, err := h.buildSummaryText(month)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	admin := auth.CurrentAdmin(r)
	pages.SummaryPreview(admin.LanguagePref, text).Render(r.Context(), w)
}

func (h *authHandlers) buildSummaryText(month string) (string, error) {
	gs, err := settings.Get(h.conn)
	if err != nil {
		return "", err
	}

	tMonth, err := time.Parse("2006-01", month)
	if err != nil {
		return "", err
	}
	monthLabel := tMonth.Format("January 2006")

	// Calculate totals for this month
	var totalCollected int64
	err = h.conn.QueryRow(
		`SELECT COALESCE(SUM(amount_minor), 0) FROM contributions WHERE contribution_month = ? AND deleted_at IS NULL`,
		month,
	).Scan(&totalCollected)
	if err != nil {
		return "", err
	}

	var totalSpent int64
	err = h.conn.QueryRow(
		`SELECT COALESCE(SUM(amount_minor), 0) FROM expenses WHERE expense_date LIKE ? || '-%' AND deleted_at IS NULL`,
		month,
	).Scan(&totalSpent)
	if err != nil {
		return "", err
	}

	// Calculate overall closing balance
	var allTimeCollected int64
	var allTimeExpenses int64
	err = h.conn.QueryRow(`SELECT COALESCE(SUM(amount_minor), 0) FROM contributions WHERE deleted_at IS NULL`).Scan(&allTimeCollected)
	if err != nil {
		return "", err
	}
	err = h.conn.QueryRow(`SELECT COALESCE(SUM(amount_minor), 0) FROM expenses WHERE deleted_at IS NULL`).Scan(&allTimeExpenses)
	if err != nil {
		return "", err
	}
	closingBalance := allTimeCollected - allTimeExpenses

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*Sadqa Ledger — %s*\n", monthLabel))
	sb.WriteString(fmt.Sprintf("Group: %s\n\n", gs.GroupName))

	sb.WriteString(fmt.Sprintf("*Monthly Summary (%s):*\n", monthLabel))
	sb.WriteString(fmt.Sprintf("- Total Collected: %s\n", formatMoney(totalCollected, gs.CurrencySymbol)))
	sb.WriteString(fmt.Sprintf("- Total Expenses: %s\n", formatMoney(totalSpent, gs.CurrencySymbol)))
	sb.WriteString(fmt.Sprintf("- Closing Balance: %s\n", formatMoney(closingBalance, gs.CurrencySymbol)))

	// Fetch all individual contributions for this month
	rows, err := h.conn.Query(
		`SELECT m.name, c.amount_minor 
		 FROM contributions c 
		 JOIN members m ON c.member_id = m.id 
		 WHERE c.contribution_month = ? AND c.deleted_at IS NULL 
		 ORDER BY c.created_at ASC, m.name COLLATE NOCASE`,
		month,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var hasContribs bool
	for rows.Next() {
		if !hasContribs {
			sb.WriteString(fmt.Sprintf("\n*Contributions (%s):*\n", monthLabel))
			hasContribs = true
		}
		var name string
		var amount int64
		if err := rows.Scan(&name, &amount); err == nil {
			displayName := name
			if !gs.ShowNamesPublicly {
				displayName = "Contributor"
			}
			sb.WriteString(fmt.Sprintf("- %s: %s\n", displayName, formatMoney(amount, gs.CurrencySymbol)))
		}
	}

	if !hasContribs {
		sb.WriteString(fmt.Sprintf("\n*Contributions (%s):*\nNo contributions recorded for %s.\n", monthLabel, monthLabel))
	}

	return sb.String(), nil
}
