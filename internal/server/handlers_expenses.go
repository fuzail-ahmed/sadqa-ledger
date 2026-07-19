package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/fuzail-ahmed/sadqa-ledger/i18n"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/auth"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/expenses"
	"github.com/fuzail-ahmed/sadqa-ledger/internal/settings"
	"github.com/fuzail-ahmed/sadqa-ledger/web/templates/pages"
)

// handleExpensesPage renders the list of expenses.
func (h *authHandlers) handleExpensesPage(w http.ResponseWriter, r *http.Request) {
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	list, err := expenses.ListActive(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	toast := r.URL.Query().Get("toast")
	pages.Expenses(h.shellData(w, r, "expenses", "nav.expenses"), list, gs.CurrencySymbol, toast).Render(r.Context(), w)
}

// handleExpenseNewPage renders the add expense form.
func (h *authHandlers) handleExpenseNewPage(w http.ResponseWriter, r *http.Request) {
	gs, err := settings.Get(h.conn)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	today := time.Now().Format("2006-01-02")
	pages.ExpenseForm(
		h.shellData(w, r, "expenses", "expenses.new_title"),
		gs.CurrencySymbol,
		"",
		"",
		today,
		pages.ExpenseFormErrors{},
	).Render(r.Context(), w)
}

// handleExpenseNewSubmit validates and saves the expense, including photo upload.
func (h *authHandlers) handleExpenseNewSubmit(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with a 10MB limit
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
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

	description := strings.TrimSpace(r.FormValue("description"))
	amountStr := r.FormValue("amount")
	expenseDate := strings.TrimSpace(r.FormValue("expense_date"))

	var errs pages.ExpenseFormErrors

	if description == "" {
		errs.Description = i18n.T(admin.LanguagePref, "validation.required")
	}

	amountMinor, amountErr := parseAmount(amountStr)
	if amountErr != nil {
		errs.Amount = "Enter an amount greater than zero."
	}

	if expenseDate == "" {
		expenseDate = time.Now().Format("2006-01-02")
	}

	// Retrieve file upload
	var receiptPhotoPath *string
	file, header, fileErr := r.FormFile("receipt_photo")
	if fileErr == nil {
		defer file.Close()

		// Validate size < 5MB
		if header.Size > 5*1024*1024 {
			errs.Photo = "Photo must be under 5MB."
		}

		// Validate type is JPG/PNG
		contentType := header.Header.Get("Content-Type")
		if contentType != "image/jpeg" && contentType != "image/jpg" && contentType != "image/png" {
			errs.Photo = "Photo must be JPG or PNG."
		}

		if errs.Photo == "" {
			// Save file locally under ./uploads
			uploadDir := "uploads"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				http.Error(w, "failed to create upload directory", http.StatusInternalServerError)
				return
			}

			// Generate unique safe name
			ext := filepath.Ext(header.Filename)
			if ext == "" {
				if contentType == "image/png" {
					ext = ".png"
				} else {
					ext = ".jpg"
				}
			}
			uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			fullPath := filepath.Join(uploadDir, uniqueName)

			out, createErr := os.Create(fullPath)
			if createErr != nil {
				http.Error(w, "failed to save photo", http.StatusInternalServerError)
				return
			}
			defer out.Close()

			if _, copyErr := io.Copy(out, file); copyErr != nil {
				http.Error(w, "failed to save photo content", http.StatusInternalServerError)
				return
			}

			// Save relative path using forward slashes
			savedPath := "uploads/" + uniqueName
			receiptPhotoPath = &savedPath
		}
	} else if fileErr != http.ErrMissingFile {
		errs.Photo = "Error reading photo file."
	}

	if errs.Description != "" || errs.Amount != "" || errs.Photo != "" {
		// Clean up the uploaded photo if we created it but validation failed elsewhere
		if receiptPhotoPath != nil {
			os.Remove(*receiptPhotoPath)
		}

		pages.ExpenseForm(
			h.shellData(w, r, "expenses", "expenses.new_title"),
			gs.CurrencySymbol,
			description,
			amountStr,
			expenseDate,
			errs,
		).Render(r.Context(), w)
		return
	}

	// Insert into DB
	_, err = expenses.Create(h.conn, description, amountMinor, expenseDate, receiptPhotoPath, admin.ID)
	if err != nil {
		if receiptPhotoPath != nil {
			os.Remove(*receiptPhotoPath)
		}
		http.Error(w, "failed to record expense", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/expenses?toast=Expense recorded successfully", http.StatusSeeOther)
}

// handleExpenseDelete handles soft deletion of an expense.
func (h *authHandlers) handleExpenseDelete(w http.ResponseWriter, r *http.Request) {
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
	if err := expenses.SoftDelete(h.conn, id, admin.ID); err != nil {
		if err == expenses.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/expenses?toast=Expense deleted successfully", http.StatusSeeOther)
}
