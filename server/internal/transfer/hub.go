// Package transfer manages the relay of a file from the phone to the PC via
// WebSocket. Each session has at most one waiting receiver (the PC).
package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const writeDeadline = 30 * time.Second

// Meta describes the file being delivered. It is sent as a JSON text frame
// immediately before the binary frame.
type Meta struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// metaMessage is the wire format of the metadata frame.
type metaMessage struct {
	Type string `json:"type"`
	Meta
}

// doneMessage is the wire format of the trailing completion frame.
type doneMessage struct {
	Type string `json:"type"`
}

// ErrNoReceiver is returned by Deliver when no PC is connected for a session.
var ErrNoReceiver = errors.New("no receiver connected")

// receiver wraps a connected WebSocket and serializes writes to it.
type receiver struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (r *receiver) close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

// Hub tracks the waiting receiver for each session.
type Hub struct {
	mu        sync.Mutex
	receivers map[string]*receiver
}

// NewHub returns an empty Hub.
func NewHub() *Hub {
	return &Hub{receivers: make(map[string]*receiver)}
}

// Register associates conn with sessionID, replacing any previous receiver.
func (h *Hub) Register(sessionID string, conn *websocket.Conn) {
	h.mu.Lock()
	if prev, ok := h.receivers[sessionID]; ok {
		prev.close()
	}
	h.receivers[sessionID] = &receiver{conn: conn}
	h.mu.Unlock()
}

// Unregister removes and closes the receiver for sessionID (if any).
func (h *Hub) Unregister(sessionID string) {
	h.mu.Lock()
	if r, ok := h.receivers[sessionID]; ok {
		r.close()
		delete(h.receivers, sessionID)
	}
	h.mu.Unlock()
}

// Deliver sends meta then data to the session's receiver as two frames
// (metadata JSON, then binary), followed by a "done" text frame.
func (h *Hub) Deliver(ctx context.Context, sessionID string, meta Meta, data []byte) error {
	h.mu.Lock()
	r, ok := h.receivers[sessionID]
	h.mu.Unlock()
	if !ok {
		return ErrNoReceiver
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.conn.SetWriteDeadline(time.Now().Add(writeDeadline)); err != nil {
		return err
	}

	metaJSON, err := json.Marshal(metaMessage{Type: "meta", Meta: meta})
	if err != nil {
		return err
	}
	if err := r.conn.WriteMessage(websocket.TextMessage, metaJSON); err != nil {
		return err
	}
	if err := r.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}
	if err := r.conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"done"}`)); err != nil {
		return err
	}

	return nil
}
