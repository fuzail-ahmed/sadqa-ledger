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

var monthNamesHindi = map[time.Month]string{
	time.January:   "जनवरी",
	time.February:  "फरवरी",
	time.March:     "मार्च",
	time.April:     "अप्रैल",
	time.May:       "मई",
	time.June:      "जून",
	time.July:      "जुलाई",
	time.August:    "अगस्त",
	time.September: "सितंबर",
	time.October:   "अक्टूबर",
	time.November:  "नवंबर",
	time.December:  "दिसंबर",
}

var monthNamesArabic = map[time.Month]string{
	time.January:   "يناير",
	time.February:  "فبراير",
	time.March:     "مارس",
	time.April:     "أبريل",
	time.May:       "مايو",
	time.June:      "يونيو",
	time.July:      "يوليو",
	time.August:    "أغسطس",
	time.September: "سبتمبر",
	time.October:   "أكتوبر",
	time.November:  "نوفمبر",
	time.December:  "ديسمبر",
}

func formatMonthLabel(t time.Time, lang string) string {
	switch lang {
	case "hi":
		if name, ok := monthNamesHindi[t.Month()]; ok {
			return fmt.Sprintf("%s %d", name, t.Year())
		}
	case "ar":
		if name, ok := monthNamesArabic[t.Month()]; ok {
			return fmt.Sprintf("%s %d", name, t.Year())
		}
	}
	return t.Format("January 2006")
}

// handleSummaryPage renders the main monthly summary screen.
func (h *authHandlers) handleSummaryPage(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	currentMonth := time.Now().Format("2006-01")
	text, err := h.buildSummaryText(currentMonth, admin.LanguagePref)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pages.Summary(h.shellData(w, r, "more", "nav.summary"), currentMonth, text).Render(r.Context(), w)
}

// handleSummaryGenerate serves HTMX requests to generate summary text for a chosen month.
func (h *authHandlers) handleSummaryGenerate(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	month := strings.TrimSpace(r.URL.Query().Get("summary_month"))
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	text, err := h.buildSummaryText(month, admin.LanguagePref)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pages.SummaryPreview(admin.LanguagePref, text).Render(r.Context(), w)
}

func (h *authHandlers) buildSummaryText(month, lang string) (string, error) {
	gs, err := settings.Get(h.conn)
	if err != nil {
		return "", err
	}

	tMonth, err := time.Parse("2006-01", month)
	if err != nil {
		return "", err
	}
	monthLabel := formatMonthLabel(tMonth, lang)

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
	switch lang {
	case "hi":
		sb.WriteString(fmt.Sprintf("*सदका लेजर — %s*\n", monthLabel))
		sb.WriteString(fmt.Sprintf("समूह: %s\n\n", gs.GroupName))

		sb.WriteString(fmt.Sprintf("*मासिक सारांश (%s):*\n", monthLabel))
		sb.WriteString(fmt.Sprintf("- कुल जमा: %s\n", formatMoney(totalCollected, gs.CurrencySymbol)))
		sb.WriteString(fmt.Sprintf("- कुल खर्च: %s\n", formatMoney(totalSpent, gs.CurrencySymbol)))
		sb.WriteString(fmt.Sprintf("- अंतिम बाकी: %s\n", formatMoney(closingBalance, gs.CurrencySymbol)))
	case "ar":
		sb.WriteString(fmt.Sprintf("*دفتر الصدقة — %s*\n", monthLabel))
		sb.WriteString(fmt.Sprintf("المجموعة: %s\n\n", gs.GroupName))

		sb.WriteString(fmt.Sprintf("*الملخص الشهري (%s):*\n", monthLabel))
		sb.WriteString(fmt.Sprintf("- إجمالي المقبوضات: %s\n", formatMoney(totalCollected, gs.CurrencySymbol)))
		sb.WriteString(fmt.Sprintf("- إجمالي المصروفات: %s\n", formatMoney(totalSpent, gs.CurrencySymbol)))
		sb.WriteString(fmt.Sprintf("- الرصيد الختامي: %s\n", formatMoney(closingBalance, gs.CurrencySymbol)))
	default:
		sb.WriteString(fmt.Sprintf("*Sadqa Ledger — %s*\n", monthLabel))
		sb.WriteString(fmt.Sprintf("Group: %s\n\n", gs.GroupName))

		sb.WriteString(fmt.Sprintf("*Monthly Summary (%s):*\n", monthLabel))
		sb.WriteString(fmt.Sprintf("- Total Collected: %s\n", formatMoney(totalCollected, gs.CurrencySymbol)))
		sb.WriteString(fmt.Sprintf("- Total Expenses: %s\n", formatMoney(totalSpent, gs.CurrencySymbol)))
		sb.WriteString(fmt.Sprintf("- Closing Balance: %s\n", formatMoney(closingBalance, gs.CurrencySymbol)))
	}

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
			switch lang {
			case "hi":
				sb.WriteString(fmt.Sprintf("\n*योगदान (%s):*\n", monthLabel))
			case "ar":
				sb.WriteString(fmt.Sprintf("\n*المساهمات (%s):*\n", monthLabel))
			default:
				sb.WriteString(fmt.Sprintf("\n*Contributions (%s):*\n", monthLabel))
			}
			hasContribs = true
		}
		var name string
		var amount int64
		if err := rows.Scan(&name, &amount); err == nil {
			displayName := name
			if !gs.ShowNamesPublicly {
				switch lang {
				case "hi":
					displayName = "योगदानकर्ता"
				case "ar":
					displayName = "مساهم"
				default:
					displayName = "Contributor"
				}
			}
			sb.WriteString(fmt.Sprintf("- %s: %s\n", displayName, formatMoney(amount, gs.CurrencySymbol)))
		}
	}

	if !hasContribs {
		switch lang {
		case "hi":
			sb.WriteString(fmt.Sprintf("\n*योगदान (%s):*\n%s के लिए कोई योगदान दर्ज नहीं किया गया।\n", monthLabel, monthLabel))
		case "ar":
			sb.WriteString(fmt.Sprintf("\n*المساهمات (%s):*\nلم يتم تسجيل أي مساهمات لشهر %s.\n", monthLabel, monthLabel))
		default:
			sb.WriteString(fmt.Sprintf("\n*Contributions (%s):*\nNo contributions recorded for %s.\n", monthLabel, monthLabel))
		}
	}

	return sb.String(), nil
}
