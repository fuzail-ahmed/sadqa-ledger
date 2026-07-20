package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/fuzail-ahmed/sadqa-ledger/i18n"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/members"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

// handleMembersPage renders the member list (docs/APP_FLOW.md §4). An
// HX-Request (live search keystroke) gets just the MemberList fragment back;
// a normal navigation gets the full page.
func (h *authHandlers) handleMembersPage(w http.ResponseWriter, r *http.Request) {
	admin := auth.CurrentAdmin(r)
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	list, err := members.List(h.conn, query)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		pages.MemberList(admin.LanguagePref, auth.CSRFToken(w, r, h.cfg.SecureCookies), query, list).Render(r.Context(), w)
		return
	}

	var toast string
	switch r.URL.Query().Get("toast") {
	case "added":
		toast = i18n.T(admin.LanguagePref, "members.added_toast")
	case "updated":
		toast = i18n.T(admin.LanguagePref, "members.updated_toast")
	}
	pages.Members(h.shellData(w, r, "members", "nav.members"), list, query, toast).Render(r.Context(), w)
}

// handleMemberNewPage renders the add-member form.
func (h *authHandlers) handleMemberNewPage(w http.ResponseWriter, r *http.Request) {
	pages.MemberForm(h.shellData(w, r, "members", "members.new_title"), "/members/new", "members.new_title", "", true, pages.MemberFormErrors{}).Render(r.Context(), w)
}

// handleMemberNewSubmit validates and creates a member, recording the acting
// admin (docs/PRD.md §5 audit trail), then redirects to the list with a
// success toast (docs/APP_FLOW.md §4a).
func (h *authHandlers) handleMemberNewSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	name := strings.TrimSpace(r.FormValue("name"))
	isActive := r.FormValue("is_active") != "0"

	if name == "" {
		errs := pages.MemberFormErrors{Name: i18n.T(admin.LanguagePref, "validation.required")}
		pages.MemberForm(h.shellData(w, r, "members", "members.new_title"), "/members/new", "members.new_title", name, isActive, errs).Render(r.Context(), w)
		return
	}

	if _, err := members.Create(h.conn, name, isActive, admin.ID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/members?toast=added", http.StatusSeeOther)
}

// handleMemberEditPage renders the edit form pre-filled with the current
// member's name and status.
func (h *authHandlers) handleMemberEditPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	m, err := members.Get(h.conn, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	action := "/members/" + chi.URLParam(r, "id") + "/edit"
	pages.MemberForm(h.shellData(w, r, "members", "members.edit_title"), action, "members.edit_title", m.Name, m.IsActive, pages.MemberFormErrors{}).Render(r.Context(), w)
}

// handleMemberEditSubmit validates and saves changes to name/status,
// recording the acting admin, then redirects to the list with a toast.
func (h *authHandlers) handleMemberEditSubmit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	name := strings.TrimSpace(r.FormValue("name"))
	isActive := r.FormValue("is_active") != "0"
	action := "/members/" + chi.URLParam(r, "id") + "/edit"

	if name == "" {
		errs := pages.MemberFormErrors{Name: i18n.T(admin.LanguagePref, "validation.required")}
		pages.MemberForm(h.shellData(w, r, "members", "members.edit_title"), action, "members.edit_title", name, isActive, errs).Render(r.Context(), w)
		return
	}

	if err := members.Update(h.conn, id, name, isActive, admin.ID); err != nil {
		if err == members.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/members?toast=updated", http.StatusSeeOther)
}

// handleMemberToggle flips a member's active flag — the Deactivate/Reactivate
// quick action on the list row (docs/APP_FLOW.md §4). This is the members
// table's soft delete: is_active, never a row deletion (docs/SCHEMA.md §3.2).
func (h *authHandlers) handleMemberToggle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := auth.VerifyCSRF(r); err != nil {
		http.Error(w, "invalid form submission, please try again", http.StatusForbidden)
		return
	}

	admin := auth.CurrentAdmin(r)
	m, err := members.Get(h.conn, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := members.SetActive(h.conn, id, !m.IsActive, admin.ID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/members", http.StatusSeeOther)
}
