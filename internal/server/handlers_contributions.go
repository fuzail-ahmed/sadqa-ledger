package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/fuzail-ahmed/sadqa-ledger/i18n"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/contributions"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/members"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/settings"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

// handleContributionNewPage renders the new contribution form.
func (h *authHandlers) handleContributionNewPage(w http.ResponseWriter, r *http.Request) {
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	currentMonth := time.Now().Format("2006-01")
	currentDate := time.Now().Format("2006-01-02")

	var selectedMember *members.Member
	if memberIDStr := r.URL.Query().Get("member_id"); memberIDStr != "" {
		if id, err := strconv.ParseInt(memberIDStr, 10, 64); err == nil {
			if m, err := members.Get(h.conn, id); err == nil && m.IsActive {
				selectedMember = m
			}
		}
	}

	pages.ContributionNew(
		h.shellData(w, r, "add", "nav.add_contribution"),
		gs.QuickAmountsMinor,
		gs.CurrencySymbol,
		currentMonth,
		currentDate,
		selectedMember,
		"",
		pages.ContributionFormErrors{},
		"",
	).Render(r.Context(), w)
}

// handleContributionSearchMembers searches active members by query prefix/substring.
func (h *authHandlers) handleContributionSearchMembers(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	query := strings.TrimSpace(r.URL.Query().Get("member_search"))

	// Retrieve members using the members package
	list, err := members.List(h.conn, query)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Filter to only active members for contribution logging
	var activeMembers []members.Member
	for _, m := range list {
		if m.IsActive {
			activeMembers = append(activeMembers, m)
		}
	}

	pages.ContributionSearchResults(admin.LanguagePref, activeMembers, time.Now().Format("2006-01")).Render(r.Context(), w)
}

// handleContributionSelectMember renders the selected member block.
func (h *authHandlers) handleContributionSelectMember(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid member id", http.StatusBadRequest)
		return
	}

	m, err := members.Get(h.conn, id)
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	month := strings.TrimSpace(r.URL.Query().Get("contribution_month"))
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	hasDuplicate, err := contributions.CheckDuplicate(h.conn, m.ID, month)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the selected member container HTML
	fmt.Fprintf(w, `<div id="member-search-container" class="field">
		<label class="label">%s</label>
		<div class="flex min-h-11 items-center justify-between gap-2 rounded-md border border-border bg-muted p-2">
			<span class="font-medium text-foreground">%s</span>
			<input type="hidden" name="member_id" value="%d"/>
			<button
				type="button"
				class="btn"
				data-size="sm"
				data-variant="outline"
				hx-get="/contributions/clear-member"
				hx-target="#member-search-container"
				hx-swap="outerHTML"
			>
				%s
			</button>
		</div>
	</div>`,
		i18n.T(admin.LanguagePref, "members.name_label"),
		m.Name,
		m.ID,
		i18n.T(admin.LanguagePref, "members.cancel"),
	)

	// Render duplicate warning out-of-band
	warningHTML := ""
	if hasDuplicate {
		warningHTML = fmt.Sprintf(`<div id="duplicate-warning" hx-swap-oob="true">
			<div class="alert text-sm" data-variant="warning" role="alert">
				<p>%s already has a payment logged for %s — this will be added as a second entry.</p>
			</div>
		</div>`, m.Name, month)
	} else {
		warningHTML = `<div id="duplicate-warning" hx-swap-oob="true"></div>`
	}
	w.Write([]byte(warningHTML))
}

// handleContributionClearMember swaps the selected member block back to a search input.
func (h *authHandlers) handleContributionClearMember(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
		<input
			class="input min-h-11"
			type="search"
			id="member_search"
			name="member_search"
			placeholder="%s"
			hx-get="/contributions/search-members"
			hx-trigger="input changed delay:200ms, search"
			hx-target="#member-search-results"
		/>
		<div id="member-search-results" class="mt-1 max-h-48 overflow-y-auto rounded border border-border empty:border-none empty:p-0 bg-surface"></div>
	`, i18n.T(admin.LanguagePref, "members.search_placeholder"))

	// Clear warning out-of-band
	w.Write([]byte(`<div id="duplicate-warning" hx-swap-oob="true"></div>`))
}

// handleContributionCheckDuplicate checks for duplicates and returns warning HTML.
func (h *authHandlers) handleContributionCheckDuplicate(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	memberIDStr := r.URL.Query().Get("member_id")
	month := strings.TrimSpace(r.URL.Query().Get("contribution_month"))

	if memberIDStr == "" || month == "" {
		w.Write([]byte(""))
		return
	}

	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		w.Write([]byte(""))
		return
	}

	m, err := members.Get(h.conn, memberID)
	if err != nil {
		w.Write([]byte(""))
		return
	}

	hasDuplicate, err := contributions.CheckDuplicate(h.conn, memberID, month)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pages.ContributionDuplicateCheck(admin.LanguagePref, hasDuplicate, m.Name, month).Render(r.Context(), w)
}

// parseAmount parses rupees/dollars into minor units (paise/cents).
func parseAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty amount")
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil || val <= 0 {
		return 0, errors.New("invalid amount")
	}
	// Convert to minor units and round to prevent floating point inaccuracies
	minor := int64(val*100 + 0.5)
	return minor, nil
}

// handleContributionNewSubmit handles submission of the contribution.
func (h *authHandlers) handleContributionNewSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	memberIDStr := r.FormValue("member_id")
	amountStr := r.FormValue("amount")
	month := strings.TrimSpace(r.FormValue("contribution_month"))
	paidOn := strings.TrimSpace(r.FormValue("paid_on"))

	var errs pages.ContributionFormErrors
	var memberID int64
	var selectedMember *members.Member

	if memberIDStr == "" {
		errs.Member = "Select a member."
	} else {
		var idErr error
		memberID, idErr = strconv.ParseInt(memberIDStr, 10, 64)
		if idErr != nil {
			errs.Member = "Invalid member selected."
		} else {
			m, getErr := members.Get(h.conn, memberID)
			if getErr != nil {
				errs.Member = "Member not found."
			} else if !m.IsActive {
				errs.Member = "Selected member is inactive."
			} else {
				selectedMember = m
			}
		}
	}

	amountMinor, amountErr := parseAmount(amountStr)
	if amountErr != nil {
		errs.Amount = "Enter an amount greater than zero."
	}

	if month == "" {
		month = time.Now().Format("2006-01")
	}
	if paidOn == "" {
		paidOn = time.Now().Format("2006-01-02")
	}

	if errs.Member != "" || errs.Amount != "" {
		// Render form contents with errors inline
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		pages.ContributionFormContent(
			admin.LanguagePref,
			gs.QuickAmountsMinor,
			gs.CurrencySymbol,
			month,
			paidOn,
			selectedMember,
			amountStr,
			errs,
		).Render(r.Context(), w)
		return
	}

	// Insert contribution
	_, err = contributions.Create(h.conn, memberID, amountMinor, month, paidOn, admin.ID)
	if err != nil {
		http.Error(w, "failed to record contribution", http.StatusInternalServerError)
		return
	}

	successMsg := fmt.Sprintf("Saved — %s%s from %s for %s", gs.CurrencySymbol, amountStr, selectedMember.Name, month)

	// If HTMX, reset the form and return content + out-of-band toast success
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Reset form contents
		pages.ContributionFormContent(
			admin.LanguagePref,
			gs.QuickAmountsMinor,
			gs.CurrencySymbol,
			time.Now().Format("2006-01"),
			time.Now().Format("2006-01-02"),
			nil, // reset selected member
			"",  // reset amount
			pages.ContributionFormErrors{},
		).Render(r.Context(), w)

		// Output-of-band toast success message
		toastHTML := fmt.Sprintf(`<div id="toast-container" hx-swap-oob="true">
			<div class="alert mb-4" data-variant="success" role="status" aria-live="polite">
				<p>%s</p>
			</div>
		</div>`, successMsg)
		w.Write([]byte(toastHTML))
		return
	}

	// Normal redirect back with toast message
	http.Redirect(w, r, fmt.Sprintf("/contributions/new?toast=%s", successMsg), http.StatusSeeOther)
}

// handleContributionsPage renders the list of contributions.
func (h *authHandlers) handleContributionsPage(w http.ResponseWriter, r *http.Request) {
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	month := strings.TrimSpace(r.URL.Query().Get("month"))
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	list, err := contributions.ListActiveByMonth(h.conn, month)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	toast := r.URL.Query().Get("toast")
	pages.Contributions(
		h.shellData(w, r, "contributions", "nav.contributions"),
		list,
		gs.CurrencySymbol,
		month,
		toast,
	).Render(r.Context(), w)
}

// handleContributionDelete soft-deletes a contribution.
func (h *authHandlers) handleContributionDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid contribution id", http.StatusBadRequest)
		return
	}

	// Fetch details for context before deleting
	c, err := contributions.Get(h.conn, id)
	if err != nil {
		http.Error(w, "contribution not found", http.StatusNotFound)
		return
	}

	err = contributions.SoftDelete(h.conn, id, admin.ID)
	if err != nil {
		http.Error(w, "failed to delete contribution", http.StatusInternalServerError)
		return
	}

	gs, _ := settings.Get(h.conn)
	amountStr := fmt.Sprintf("%.2f", float64(c.AmountMinor)/100.0)
	if c.AmountMinor%100 == 0 {
		amountStr = fmt.Sprintf("%d", c.AmountMinor/100)
	}
	msg := fmt.Sprintf("Deleted contribution entry — %s%s from %s for %s", gs.CurrencySymbol, amountStr, c.MemberName, c.ContributionMonth)

	http.Redirect(w, r, fmt.Sprintf("/contributions?month=%s&toast=%s", c.ContributionMonth, msg), http.StatusSeeOther)
}
