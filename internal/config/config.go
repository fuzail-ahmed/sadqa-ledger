// Package config loads runtime configuration for the server from environment
// variables. Phase 0 only needs the listen port; later phases will extend
// this with database path, session secret, etc. (see .env.example).
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds server configuration read from the environment.
type Config struct {
	// Port the HTTP server listens on. Defaults to 8080, matching
	// .env.example's PORT value.
	Port int

	// DatabasePath is the path to the SQLite database file. Defaults to
	// sadqa-ledger.db in the working directory for local dev; .env.example's
	// DATABASE_PATH points inside a mounted volume for Docker deployments.
	DatabasePath string

	// SessionLifetime is how long an admin stays logged in before needing to
	// log in again (docs/SCHEMA.md §5), refreshed on activity. Sourced from
	// SESSION_LIFETIME_DAYS, default 30 per .env.example.
	SessionLifetime time.Duration

	// SecureCookies controls the session/CSRF cookies' Secure flag. Derived
	// from BASE_URL's scheme rather than a separate env var: true for
	// https:// deployments, false only when BASE_URL is explicitly
	// http://... for local dev (docs/SCHEMA.md §5's "configurable off for
	// local HTTP dev" requirement).
	SecureCookies bool

	// DefaultCurrencyCode/DefaultCurrencySymbol prefill the first-run setup
	// wizard (docs/SCHEMA.md §8 seed-data note); an admin can change them
	// later from Settings.
	DefaultCurrencyCode   string
	DefaultCurrencySymbol string
}

// Load reads configuration from environment variables, falling back to
// sane defaults for local development so the server can start even before
// .env is populated.
func Load() Config {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			port = p
		}
	}

	dbPath := "sadqa-ledger.db"
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		dbPath = v
	}

	lifetimeDays := 30
	if v := os.Getenv("SESSION_LIFETIME_DAYS"); v != "" {
		if d, err := strconv.Atoi(v); err == nil && d > 0 {
			lifetimeDays = d
		}
	}

	secureCookies := !strings.HasPrefix(os.Getenv("BASE_URL"), "http://")

	currencyCode := "INR"
	if v := os.Getenv("DEFAULT_CURRENCY_CODE"); v != "" {
		currencyCode = v
	}
	currencySymbol := "₹"
	if v := os.Getenv("DEFAULT_CURRENCY_SYMBOL"); v != "" {
		currencySymbol = v
	}

	return Config{
		Port:                  port,
		DatabasePath:          dbPath,
		SessionLifetime:       time.Duration(lifetimeDays) * 24 * time.Hour,
		SecureCookies:         secureCookies,
		DefaultCurrencyCode:   currencyCode,
		DefaultCurrencySymbol: currencySymbol,
	}
}
