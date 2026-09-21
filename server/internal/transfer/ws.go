package transfer

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader allows any origin. The mobile page is same-origin and the extension
// connects over host permissions, but the origin check is left permissive so
// local/off-domain development works without extra configuration.
var upgrader = websocket.Upgrader{
	CheckOrigin:     func(*http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
}

// Upgrade performs the WebSocket handshake and registers conn as the receiver
// for sessionID. On success it returns the connection so the caller can run
// the read loop.
func (h *Hub) Upgrade(w http.ResponseWriter, r *http.Request, sessionID string) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	h.Register(sessionID, conn)
	return conn, nil
}

// HandleConn reads frames from the receiver until it disconnects or cancels,
// then unregisters it. It must be run in its own goroutine.
func (h *Hub) HandleConn(ctx context.Context, sessionID string, conn *websocket.Conn, log *slog.Logger) {
	defer func() {
		h.Unregister(sessionID)
		_ = conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		// The receiver may send a cancel to release the session explicitly.
		if string(msg) == `{"type":"cancel"}` {
			return
		}
		// Ignore any other inbound payload; this direction is receive-only.
		_ = ctx
	}
}
