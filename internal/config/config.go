// Package config loads runtime configuration for the server from environment
// variables. Phase 0 only needs the listen port; later phases will extend
// this with database path, session secret, etc. (see .env.example).
package config

import (
	"os"
	"strconv"
)

// Config holds server configuration read from the environment.
type Config struct {
	// Port the HTTP server listens on. Defaults to 8080, matching
	// .env.example's PORT value.
	Port int
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
	return Config{Port: port}
}
