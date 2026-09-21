package session

import (
	"errors"
	"testing"
	"time"
)

func TestCreateAndGet(t *testing.T) {
	s := NewStore(time.Minute)
	sess, err := s.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := s.Get(sess.ID, sess.Token)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != sess.ID {
		t.Errorf("got ID %q, want %q", got.ID, sess.ID)
	}
}

func TestGetWrongToken(t *testing.T) {
	s := NewStore(time.Minute)
	sess, _ := s.Create()

	if _, err := s.Get(sess.ID, "deadbeefdeadbeefdeadbeefdeadbeef"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Get() error = %v, want ErrInvalidToken", err)
	}
}

func TestGetNotFound(t *testing.T) {
	s := NewStore(time.Minute)
	if _, err := s.Get("missing", "token"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestGetExpired(t *testing.T) {
	s := NewStore(time.Minute)
	sess, _ := s.Create()

	// Manually force the session to be expired.
	s.mu.Lock()
	s.sessions[sess.ID] = Session{
		ID:        sess.ID,
		Token:     sess.Token,
		CreatedAt: sess.CreatedAt,
		ExpiresAt: time.Now().Add(-time.Second),
	}
	s.mu.Unlock()

	if _, err := s.Get(sess.ID, sess.Token); !errors.Is(err, ErrExpired) {
		t.Fatalf("Get() error = %v, want ErrExpired", err)
	}

	// The expired session must have been removed.
	if _, err := s.Get(sess.ID, sess.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() after expiry error = %v, want ErrNotFound", err)
	}
}

func TestDelete(t *testing.T) {
	s := NewStore(time.Minute)
	sess, _ := s.Create()

	s.Delete(sess.ID)
	if _, err := s.Get(sess.ID, sess.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestCleanup(t *testing.T) {
	s := NewStore(time.Minute)
	live, _ := s.Create()

	// Insert an already-expired session directly.
	expired := Session{
		ID:        "expired-id",
		Token:     "token",
		CreatedAt: time.Now().Add(-2 * time.Minute),
		ExpiresAt: time.Now().Add(-time.Minute),
	}
	s.mu.Lock()
	s.sessions[expired.ID] = expired
	s.mu.Unlock()

	s.Cleanup(time.Now())

	if _, err := s.Get(live.ID, live.Token); err != nil {
		t.Fatalf("live session removed: %v", err)
	}
	if _, err := s.Get(expired.ID, expired.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired session still present: %v", err)
	}
}

func TestUniqueIDs(t *testing.T) {
	s := NewStore(time.Minute)
	a, _ := s.Create()
	b, _ := s.Create()
	if a.ID == b.ID {
		t.Fatal("expected unique session IDs")
	}
	if a.Token == b.Token {
		t.Fatal("expected unique tokens")
	}
}
