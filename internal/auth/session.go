package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
)

// timeLayout matches the ISO 8601 UTC format docs/SCHEMA.md §1 specifies for
// all stored timestamps; used for expires_at, which (unlike created_at) has
// no SQLite-side default and is always written by the app.
const timeLayout = "2006-01-02T15:04:05.000Z"

// ErrNoSession is returned when a session token has no matching row, or the
// matching row is expired.
var ErrNoSession = errors.New("no matching session")

// Session is a row from the sessions table (docs/SCHEMA.md §3.6).
type Session struct {
	AdminID   int64
	ExpiresAt time.Time
}

// newRawToken generates a cryptographically random session token
// (docs/SCHEMA.md §5: 32 bytes from crypto/rand, encoded for cookie use).
func newRawToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// hashToken computes the SHA-256 hex digest stored as sessions.token_hash.
// The raw token itself is never written to the database (docs/SCHEMA.md §5).
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// CreateSession generates a new random token, stores its hash, and returns
// the raw token for the caller to set as the cookie value. Always called on
// login to produce a brand-new session (no fixation — the old session, if
// any, is never reused).
func CreateSession(conn *sql.DB, adminID int64, lifetime time.Duration, userAgent, ipAddress string) (rawToken string, err error) {
	rawToken, err = newRawToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().UTC().Add(lifetime).Format(timeLayout)
	_, err = conn.Exec(
		`INSERT INTO sessions (token_hash, admin_id, expires_at, user_agent, ip_address) VALUES (?, ?, ?, ?, ?)`,
		hashToken(rawToken), adminID, expiresAt, userAgent, ipAddress,
	)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

// LookupSession hashes rawToken and looks up the session (docs/SCHEMA.md §5
// lookup flow). An expired row is deleted on the way out and treated as no
// session. On success, last_seen_at is updated (sliding expiration).
func LookupSession(conn *sql.DB, rawToken string) (*Session, error) {
	hash := hashToken(rawToken)

	var adminID int64
	var expiresAtStr string
	err := conn.QueryRow(`SELECT admin_id, expires_at FROM sessions WHERE token_hash = ?`, hash).Scan(&adminID, &expiresAtStr)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoSession
	}
	if err != nil {
		return nil, err
	}

	expiresAt, err := time.Parse(timeLayout, expiresAtStr)
	if err != nil {
		return nil, err
	}
	if time.Now().UTC().After(expiresAt) {
		conn.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hash)
		return nil, ErrNoSession
	}

	now := time.Now().UTC().Format(timeLayout)
	conn.Exec(`UPDATE sessions SET last_seen_at = ? WHERE token_hash = ?`, now, hash)

	return &Session{AdminID: adminID, ExpiresAt: expiresAt}, nil
}

// DeleteSession removes the session row matching rawToken (logout —
// server-side invalidation, not just clearing the cookie).
func DeleteSession(conn *sql.DB, rawToken string) error {
	_, err := conn.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hashToken(rawToken))
	return err
}

// DeleteExpiredSessions sweeps every session past its expiry, per
// docs/SCHEMA.md §5's "periodic sweep" cleanup note. Safe to call
// repeatedly (e.g. from a ticker in cmd/server).
func DeleteExpiredSessions(conn *sql.DB) (int64, error) {
	now := time.Now().UTC().Format(timeLayout)
	res, err := conn.Exec(`DELETE FROM sessions WHERE expires_at <= ?`, now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// constantTimeEqual compares two strings without leaking timing
// information, used for CSRF token comparison.
func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
