package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
)

// Admin is a row from the admins table (docs/SCHEMA.md §3.1).
type Admin struct {
	ID           int64
	Username     string
	PasswordHash string
	DisplayName  string
	LanguagePref string
}

// ErrAdminNotFound is returned by FindActiveAdminByUsername when no active
// admin matches — callers must still run VerifyAgainstDummy to avoid a
// timing-based username-enumeration signal (docs/APP_FLOW.md §1).
var ErrAdminNotFound = errors.New("admin not found")

// FindActiveAdminByUsername looks up an active admin by username.
// Deactivated admins (is_active = 0) can't log in, per docs/SCHEMA.md §3.1.
func FindActiveAdminByUsername(conn *sql.DB, username string) (*Admin, error) {
	var a Admin
	err := conn.QueryRow(
		`SELECT id, username, password_hash, display_name, language_pref FROM admins WHERE username = ? AND is_active = 1`,
		username,
	).Scan(&a.ID, &a.Username, &a.PasswordHash, &a.DisplayName, &a.LanguagePref)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAdminNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// FindAdminByID looks up an active admin by id, used by the auth middleware
// to attach the acting admin to each authenticated request's context.
func FindAdminByID(conn *sql.DB, id int64) (*Admin, error) {
	var a Admin
	err := conn.QueryRow(
		`SELECT id, username, password_hash, display_name, language_pref FROM admins WHERE id = ? AND is_active = 1`,
		id,
	).Scan(&a.ID, &a.Username, &a.PasswordHash, &a.DisplayName, &a.LanguagePref)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAdminNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// CreateAdmin inserts a new admin row with an already-bcrypt-hashed
// password, recording which admin created it (docs/SCHEMA.md §3.1).
// createdByAdminID of 0 stores NULL (no creator, e.g. test setup). Returns
// the new admin's id.
func CreateAdmin(conn *sql.DB, username, passwordHash, displayName string, createdByAdminID int64) (int64, error) {
	var createdBy sql.NullInt64
	if createdByAdminID != 0 {
		createdBy = sql.NullInt64{Int64: createdByAdminID, Valid: true}
	}
	res, err := conn.Exec(
		`INSERT INTO admins (username, password_hash, display_name, created_by_admin_id) VALUES (?, ?, ?, ?)`,
		username, passwordHash, displayName, createdBy,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// SetLanguagePref updates an admin's UI language, used by the language
// switcher in the shared admin shell (docs/APP_FLOW.md §10).
func SetLanguagePref(conn *sql.DB, adminID int64, lang string) error {
	_, err := conn.Exec(`UPDATE admins SET language_pref = ? WHERE id = ?`, lang, adminID)
	return err
}

// GroupSettingsExist reports whether the singleton group_settings row has
// been created yet — the first-run check from docs/SCHEMA.md §8: setup is
// only reachable while this is false, and is created together with the
// first admin in the same transaction (see CreateGroupInSetup).
func GroupSettingsExist(conn *sql.DB) (bool, error) {
	var exists int
	err := conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM group_settings WHERE id = 1)`).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// newPublicToken generates the unguessable random token for /p/:token
// (docs/SCHEMA.md §3.5), 32 bytes hex-encoded.
func newPublicToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// CreateGroupInSetup performs the first-run wizard's write: the first admin
// account and the singleton group_settings row, in one transaction, with a
// freshly generated public_token (docs/SCHEMA.md §8 seed-data note).
func CreateGroupInSetup(conn *sql.DB, username, passwordHash, displayName, groupName, currencyCode, currencySymbol string) (adminID int64, err error) {
	publicToken, err := newPublicToken()
	if err != nil {
		return 0, err
	}

	tx, err := conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO admins (username, password_hash, display_name) VALUES (?, ?, ?)`,
		username, passwordHash, displayName,
	)
	if err != nil {
		return 0, err
	}
	adminID, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(
		`INSERT INTO group_settings (id, group_name, currency_code, currency_symbol, public_token) VALUES (1, ?, ?, ?, ?)`,
		groupName, currencyCode, currencySymbol, publicToken,
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return adminID, nil
}

// ListAdmins returns all admin accounts, active or not (for settings management).
func ListAdmins(conn *sql.DB) ([]Admin, error) {
	rows, err := conn.Query(`SELECT id, username, password_hash, display_name, language_pref FROM admins ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Admin
	for rows.Next() {
		var a Admin
		if err := rows.Scan(&a.ID, &a.Username, &a.PasswordHash, &a.DisplayName, &a.LanguagePref); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
