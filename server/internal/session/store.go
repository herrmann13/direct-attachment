package session

import (
	"crypto/subtle"
	"errors"
	"sync"
	"time"
)

// Sentinel errors returned by Store.Get.
var (
	ErrNotFound     = errors.New("session not found")
	ErrInvalidToken = errors.New("invalid token")
	ErrExpired      = errors.New("session expired")
)

// Store is a concurrency-safe in-memory session store. For horizontal scaling
// this could be swapped for a Redis-backed implementation implementing the same
// surface.
type Store struct {
	mu       sync.RWMutex
	sessions map[string]Session
	ttl      time.Duration
}

// NewStore returns a Store that creates sessions with the given TTL.
func NewStore(ttl time.Duration) *Store {
	return &Store{
		sessions: make(map[string]Session),
		ttl:      ttl,
	}
}

// Create generates and persists a new session.
func (s *Store) Create() (Session, error) {
	sess, err := New(s.ttl)
	if err != nil {
		return Session{}, err
	}

	s.mu.Lock()
	s.sessions[sess.ID] = sess
	s.mu.Unlock()

	return sess, nil
}

// Get looks up a session and validates both the token and the expiry. An
// expired session is removed lazily before returning ErrExpired.
func (s *Store) Get(id, token string) (Session, error) {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()

	if !ok {
		return Session{}, ErrNotFound
	}
	if sess.Expired(time.Now()) {
		s.Delete(id)
		return Session{}, ErrExpired
	}
	// Constant-time comparison to avoid leaking token contents via timing.
	if subtle.ConstantTimeCompare([]byte(sess.Token), []byte(token)) != 1 {
		return Session{}, ErrInvalidToken
	}

	return sess, nil
}

// Delete removes a session by ID. It is safe to call on a missing ID.
func (s *Store) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

// Cleanup removes every session expired at time now.
func (s *Store) Cleanup(now time.Time) {
	s.mu.Lock()
	for id, sess := range s.sessions {
		if sess.Expired(now) {
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()
}

// RunCleanup blocks and periodically runs Cleanup until stop is closed.
func (s *Store) RunCleanup(interval time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-stop:
			return
		case now := <-t.C:
			s.Cleanup(now)
		}
	}
}
