// Package settings manages the singleton group_settings configuration row
// (docs/SCHEMA.md §3.5, docs/APP_FLOW.md §7).
package settings

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
)

// GroupSettings represents the singleton configuration row from the group_settings table.
type GroupSettings struct {
	GroupName             string
	CurrencyCode          string
	CurrencySymbol        string
	ShowNamesPublicly     bool
	PublicToken           string
	QuickAmountsMinor     []int64 // JSON array parsed/marshaled on save/load
	DefaultPublicLanguage string
	PrivacyPolicyURL      *string // Nullable TEXT column, Phase 7 additions
}

var ErrNotFound = errors.New("group settings not found")

// Get retrieves the singleton settings row (id = 1).
func Get(conn *sql.DB) (*GroupSettings, error) {
	var gs GroupSettings
	var showNames int
	var quickAmounts string
	var privacyPolicy sql.NullString

	err := conn.QueryRow(
		`SELECT group_name, currency_code, currency_symbol, show_names_publicly, 
		        public_token, quick_amounts_minor, default_public_language, privacy_policy_url 
		 FROM group_settings WHERE id = 1`,
	).Scan(
		&gs.GroupName, &gs.CurrencyCode, &gs.CurrencySymbol, &showNames,
		&gs.PublicToken, &quickAmounts, &gs.DefaultPublicLanguage, &privacyPolicy,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	gs.ShowNamesPublicly = showNames == 1

	if privacyPolicy.Valid {
		gs.PrivacyPolicyURL = &privacyPolicy.String
	}

	if err := json.Unmarshal([]byte(quickAmounts), &gs.QuickAmountsMinor); err != nil {
		// Fallback to default if JSON is malformed
		gs.QuickAmountsMinor = []int64{20000, 50000, 100000, 200000}
	}

	return &gs, nil
}

// Update saves the settings back to the singleton row.
func Update(conn *sql.DB, gs *GroupSettings) error {
	quickAmountsJSON, err := json.Marshal(gs.QuickAmountsMinor)
	if err != nil {
		return err
	}

	showNames := 0
	if gs.ShowNamesPublicly {
		showNames = 1
	}

	var privacyPolicy sql.NullString
	if gs.PrivacyPolicyURL != nil {
		privacyPolicy = sql.NullString{String: *gs.PrivacyPolicyURL, Valid: true}
	}

	_, err = conn.Exec(
		`UPDATE group_settings 
		 SET group_name = ?, currency_code = ?, currency_symbol = ?, 
		     show_names_publicly = ?, quick_amounts_minor = ?, 
		     default_public_language = ?, privacy_policy_url = ?, 
		     updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') 
		 WHERE id = 1`,
		gs.GroupName, gs.CurrencyCode, gs.CurrencySymbol,
		showNames, string(quickAmountsJSON), gs.DefaultPublicLanguage, privacyPolicy,
	)
	return err
}

// RegeneratePublicToken creates a new token and updates the database, invalidating the old one.
func RegeneratePublicToken(conn *sql.DB) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	newToken := hex.EncodeToString(buf)

	_, err := conn.Exec(
		`UPDATE group_settings 
		 SET public_token = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') 
		 WHERE id = 1`,
		newToken,
	)
	if err != nil {
		return "", err
	}
	return newToken, nil
}
