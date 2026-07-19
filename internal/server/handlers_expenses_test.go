package server

import (
	"bytes"
	"database/sql"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
)

func createMultipartRequest(
	t *testing.T,
	path string,
	form url.Values,
	fieldName string,
	fileName string,
	contentType string,
	fileContent []byte,
	cookies []*http.Cookie,
) *http.Request {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for k, vs := range form {
		for _, v := range vs {
			_ = writer.WriteField(k, v)
		}
	}

	if fileContent != nil {
		h := make(map[string][]string)
		h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName)}
		h["Content-Type"] = []string{contentType}
		part, err := writer.CreatePart(h)
		if err != nil {
			t.Fatalf("CreatePart: %v", err)
		}
		_, _ = part.Write(fileContent)
	}

	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func TestExpensesFlow(t *testing.T) {
	h, conn := newTestServer(t)
	completeSetup(t, h)
	loginGet := doGet(h, "/login", nil)
	loginCsrf, cookies := extractCSRF(t, loginGet)
	loginPost := doPostForm(h, "/login", url.Values{
		"csrf_token": {loginCsrf}, "username": {"sohail"}, "password": {"correct-horse-battery"},
	}, cookies)
	sessionCookies := mergeCookies(cookies, loginPost.Result().Cookies())

	// --- 1. View empty expenses list ---
	listEmpty := doGet(h, "/expenses", sessionCookies)
	if listEmpty.Code != 200 {
		t.Fatalf("GET /expenses = %d, want 200", listEmpty.Code)
	}
	if !strings.Contains(listEmpty.Body.String(), "No expenses recorded yet.") {
		t.Error("Empty expenses list missing placeholder text")
	}

	// --- 2. Add expense form page ---
	newGet := doGet(h, "/expenses/new", sessionCookies)
	if newGet.Code != 200 {
		t.Fatalf("GET /expenses/new = %d, want 200", newGet.Code)
	}
	expenseCsrf, expenseCookies := extractCSRF(t, newGet)
	expenseSessionCookies := mergeCookies(sessionCookies, expenseCookies)

	// --- 3. Submit invalid expense (missing description) ---
	invalidReq := createMultipartRequest(
		t,
		"/expenses/new",
		url.Values{
			"csrf_token":   {expenseCsrf},
			"description":  {""},
			"amount":       {"150"},
			"expense_date": {"2026-07-19"},
		},
		"receipt_photo",
		"",
		"",
		nil,
		expenseSessionCookies,
	)
	invalidPost := httptest.NewRecorder()
	h.ServeHTTP(invalidPost, invalidReq)
	if !strings.Contains(invalidPost.Body.String(), "required") {
		t.Error("Validation error for missing description not rendered")
	}

	// --- 4. Submit invalid photo type ---
	badPhotoReq := createMultipartRequest(
		t,
		"/expenses/new",
		url.Values{
			"csrf_token":   {expenseCsrf},
			"description":  {"Electricity Bill"},
			"amount":       {"250"},
			"expense_date": {"2026-07-19"},
		},
		"receipt_photo",
		"doc.txt",
		"text/plain",
		[]byte("dummy text file content"),
		expenseSessionCookies,
	)
	recBadPhoto := httptest.NewRecorder()
	h.ServeHTTP(recBadPhoto, badPhotoReq)
	if !strings.Contains(recBadPhoto.Body.String(), "Photo must be JPG or PNG.") {
		t.Error("Incorrect photo type error message missing")
	}

	// --- 5. Submit oversized photo ---
	oversizedContent := make([]byte, 6*1024*1024) // 6MB
	badSizeReq := createMultipartRequest(
		t,
		"/expenses/new",
		url.Values{
			"csrf_token":   {expenseCsrf},
			"description":  {"Electricity Bill"},
			"amount":       {"250"},
			"expense_date": {"2026-07-19"},
		},
		"receipt_photo",
		"receipt.png",
		"image/png",
		oversizedContent,
		expenseSessionCookies,
	)
	recBadSize := httptest.NewRecorder()
	h.ServeHTTP(recBadSize, badSizeReq)
	if !strings.Contains(recBadSize.Body.String(), "Photo must be under 5MB.") {
		t.Error("Oversized photo error message missing")
	}

	// --- 6. Submit valid expense with photo ---
	validReq := createMultipartRequest(
		t,
		"/expenses/new",
		url.Values{
			"csrf_token":   {expenseCsrf},
			"description":  {"Electricity Bill"},
			"amount":       {"120.50"},
			"expense_date": {"2026-07-19"},
		},
		"receipt_photo",
		"receipt.png",
		"image/png",
		[]byte("fake png image content"),
		expenseSessionCookies,
	)
	recValid := httptest.NewRecorder()
	h.ServeHTTP(recValid, validReq)
	if recValid.Code != 303 {
		t.Fatalf("POST /expenses/new = %d, want 303 redirect", recValid.Code)
	}

	// Verify database entry
	var description string
	var amountMinor int64
	var photoPath sql.NullString
	err := conn.QueryRow(`SELECT description, amount_minor, receipt_photo_path FROM expenses WHERE description = 'Electricity Bill'`).Scan(&description, &amountMinor, &photoPath)
	if err != nil {
		t.Fatalf("Query expense: %v", err)
	}
	if amountMinor != 12050 {
		t.Errorf("amountMinor = %d, want 12050", amountMinor)
	}
	if !photoPath.Valid || !strings.HasPrefix(photoPath.String, "uploads/") {
		t.Errorf("photoPath = %v, want prefix uploads/", photoPath)
	}

	// Clean up created file
	if photoPath.Valid {
		defer os.Remove(photoPath.String)
	}

	// --- 7. List reflects new expense and alt text exists ---
	listAfterAdd := doGet(h, "/expenses", expenseSessionCookies)
	listBody := listAfterAdd.Body.String()
	if !strings.Contains(listBody, "Electricity Bill") {
		t.Error("Expenses list does not show newly added expense")
	}
	if !strings.Contains(listBody, `alt="Receipt photo for: Electricity Bill, 2026-07-19"`) {
		t.Error("Image alt text is missing or incorrect")
	}

	// --- 8. Delete expense ---
	var expenseID int64
	conn.QueryRow(`SELECT id FROM expenses LIMIT 1`).Scan(&expenseID)
	idStr := strconv.FormatInt(expenseID, 10)

	deleteReq := doGet(h, "/expenses", expenseSessionCookies) // fresh CSRF
	deleteCsrf, deleteCookies := extractCSRF(t, deleteReq)
	deleteSessionCookies := mergeCookies(expenseSessionCookies, deleteCookies)

	recDelete := doPostForm(h, "/expenses/"+idStr+"/delete", url.Values{
		"csrf_token": {deleteCsrf},
	}, deleteSessionCookies)
	if recDelete.Code != 303 {
		t.Fatalf("POST /expenses/:id/delete = %d, want 303", recDelete.Code)
	}

	var deletedAt sql.NullString
	conn.QueryRow(`SELECT deleted_at FROM expenses WHERE id = ?`, expenseID).Scan(&deletedAt)
	if !deletedAt.Valid || deletedAt.String == "" {
		t.Error("Expense was not soft-deleted")
	}
}
