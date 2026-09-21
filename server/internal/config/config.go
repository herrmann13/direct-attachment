// Package config loads and validates the server configuration from
// environment variables. Keeping this isolated makes it easy to document the
// full surface of deploy-time settings in a single place.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	// DefaultPort is used when PORT is not set.
	DefaultPort = "8080"
	// DefaultSessionTTL is how long a pairing session stays valid.
	DefaultSessionTTL = 2 * time.Minute
	// DefaultMaxFileSize caps the uploaded image size (15 MiB).
	DefaultMaxFileSize = 15 << 20
)

// Config holds every tunable knob for the server.
type Config struct {
	// Port is the HTTP listen port (no leading colon).
	Port string

	// BaseURL is the public origin used to build QR code URLs. When empty, it
	// is derived from the incoming request (Host + X-Forwarded-Proto), which
	// works transparently behind Railway's proxy.
	BaseURL string

	// SessionTTL is the lifetime of a pairing session.
	SessionTTL time.Duration

	// MaxFileSize is the maximum accepted upload size in bytes.
	MaxFileSize int64
}

// Load reads configuration from the environment and validates it.
func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", DefaultPort),
		BaseURL:     os.Getenv("BASE_URL"),
		SessionTTL:  DefaultSessionTTL,
		MaxFileSize: DefaultMaxFileSize,
	}

	if v := os.Getenv("SESSION_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SESSION_TTL %q: %w", v, err)
		}
		if d <= 0 {
			return Config{}, fmt.Errorf("invalid SESSION_TTL %q: must be positive", v)
		}
		cfg.SessionTTL = d
	}

	if v := os.Getenv("MAX_FILE_SIZE"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("invalid MAX_FILE_SIZE %q: must be a positive integer (bytes)", v)
		}
		cfg.MaxFileSize = n
	}

	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
