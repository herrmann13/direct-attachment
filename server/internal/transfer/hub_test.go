package transfer

import (
	"context"
	"errors"
	"testing"
)

func TestDeliverWithoutReceiver(t *testing.T) {
	h := NewHub()
	err := h.Deliver(context.Background(), "nope", Meta{Name: "x.png"}, []byte("data"))
	if !errors.Is(err, ErrNoReceiver) {
		t.Fatalf("Deliver() error = %v, want ErrNoReceiver", err)
	}
}

func TestRegisterUnregister(t *testing.T) {
	h := NewHub()

	// Registering a nil connection exercises the bookkeeping paths without a
	// real socket; Deliver after unregister must return ErrNoReceiver.
	h.Register("sess", nil)
	h.Unregister("sess")

	if err := h.Deliver(context.Background(), "sess", Meta{}, nil); !errors.Is(err, ErrNoReceiver) {
		t.Fatalf("Deliver() after Unregister error = %v, want ErrNoReceiver", err)
	}
}
