// Package auth implements password hashing, session tokens, CSRF
// protection, and the chi middleware that guards admin routes, per
// docs/SCHEMA.md §5 and docs/TRD.md §6/§9.
package auth

import "golang.org/x/crypto/bcrypt"

// dummyHash is compared against on a login attempt for a username that
// doesn't exist, so failed logins take roughly the same time whether the
// username is real or not (docs/APP_FLOW.md §1: no user-enumeration signal).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

// HashPassword bcrypt-hashes a plaintext password at the standard cost
// factor (docs/TRD.md §9). The plaintext is never logged or stored.
func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword reports whether plaintext matches the bcrypt hash.
// bcrypt.CompareHashAndPassword is already constant-time.
func VerifyPassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// VerifyAgainstDummy runs a bcrypt comparison against a fixed dummy hash so
// the "unknown username" path costs the same as "wrong password" — callers
// use this when no admin row was found, instead of skipping the bcrypt call.
func VerifyAgainstDummy(plaintext string) {
	bcrypt.CompareHashAndPassword(dummyHash, []byte(plaintext))
}
