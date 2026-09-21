// Package session defines the pairing session domain model and its in-memory
// store. A session pairs one PC (Chrome extension) with one phone.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Session represents a single pairing attempt between PC and phone.
type Session struct {
	ID        string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// newToken returns a cryptographically random hex token.
func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// New creates a session with a random ID and token, expiring after ttl.
func New(ttl time.Duration) (Session, error) {
	id, err := newToken()
	if err != nil {
		return Session{}, err
	}
	token, err := newToken()
	if err != nil {
		return Session{}, err
	}

	now := time.Now()
	return Session{
		ID:        id,
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}, nil
}

// Expired reports whether the session is past its lifetime at time now.
func (s Session) Expired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}
