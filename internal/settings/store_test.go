package settings

import (
	"database/sql"
	"testing"

	sqldb "github.com/fuzail-ahmed/sadqa-ledger/internal/db"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sqldb.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	schema := `
	CREATE TABLE group_settings (
		id                       INTEGER PRIMARY KEY CHECK (id = 1),
		group_name               TEXT NOT NULL,
		currency_code            TEXT NOT NULL DEFAULT 'INR',
		currency_symbol          TEXT NOT NULL DEFAULT '₹',
		show_names_publicly      INTEGER NOT NULL DEFAULT 0,
		public_token             TEXT NOT NULL UNIQUE,
		quick_amounts_minor      TEXT NOT NULL DEFAULT '[20000,50000,100000,200000]',
		default_public_language  TEXT NOT NULL DEFAULT 'en',
		privacy_policy_url       TEXT,
		updated_at               TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	);`
	if _, err := conn.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO group_settings (id, group_name, public_token) VALUES (1, 'Masjid', 'token123')`); err != nil {
		t.Fatalf("insert settings: %v", err)
	}
	return conn
}

func TestGetSettings(t *testing.T) {
	conn := openTestDB(t)

	gs, err := Get(conn)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if gs.GroupName != "Masjid" {
		t.Errorf("GroupName = %q, want %q", gs.GroupName, "Masjid")
	}
	if gs.PublicToken != "token123" {
		t.Errorf("PublicToken = %q, want %q", gs.PublicToken, "token123")
	}
	if len(gs.QuickAmountsMinor) != 4 || gs.QuickAmountsMinor[0] != 20000 {
		t.Errorf("QuickAmountsMinor = %v, want defaults", gs.QuickAmountsMinor)
	}
}

func TestUpdateSettings(t *testing.T) {
	conn := openTestDB(t)

	gs, _ := Get(conn)
	gs.GroupName = "New Masjid"
	gs.QuickAmountsMinor = []int64{10000, 30000}
	gs.ShowNamesPublicly = true
	url := "https://example.com"
	gs.PrivacyPolicyURL = &url

	if err := Update(conn, gs); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := Get(conn)
	if err != nil {
		t.Fatalf("Get updated: %v", err)
	}

	if updated.GroupName != "New Masjid" {
		t.Errorf("GroupName = %q, want %q", updated.GroupName, "New Masjid")
	}
	if len(updated.QuickAmountsMinor) != 2 || updated.QuickAmountsMinor[0] != 10000 {
		t.Errorf("QuickAmountsMinor = %v, want [10000, 30000]", updated.QuickAmountsMinor)
	}
	if !updated.ShowNamesPublicly {
		t.Error("ShowNamesPublicly is false, want true")
	}
	if updated.PrivacyPolicyURL == nil || *updated.PrivacyPolicyURL != url {
		t.Errorf("PrivacyPolicyURL = %v, want %q", updated.PrivacyPolicyURL, url)
	}
}

func TestRegeneratePublicToken(t *testing.T) {
	conn := openTestDB(t)

	oldToken := "token123"
	newToken, err := RegeneratePublicToken(conn)
	if err != nil {
		t.Fatalf("RegeneratePublicToken: %v", err)
	}

	if newToken == oldToken {
		t.Errorf("newToken is same as oldToken")
	}

	gs, _ := Get(conn)
	if gs.PublicToken != newToken {
		t.Errorf("gs.PublicToken = %q, want %q", gs.PublicToken, newToken)
	}
}
