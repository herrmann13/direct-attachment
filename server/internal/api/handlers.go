package api

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/direct-attachment-plugin/server/internal/httpx"
	"github.com/direct-attachment-plugin/server/internal/session"
	"github.com/direct-attachment-plugin/server/internal/transfer"
	"github.com/direct-attachment-plugin/server/internal/upload"
)

// handleCreateSession generates a new pairing session and returns its QR URL.
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	sess, err := s.store.Create()
	if err != nil {
		s.log.Error("failed to create session", "error", err)
		httpx.WriteError(w, httpx.Internal("failed to create session"))
		return
	}

	url := s.buildURL(r, "/s/"+sess.ID+"?t="+sess.Token)

	httpx.WriteJSON(w, http.StatusCreated, map[string]string{
		"sessionId": sess.ID,
		"token":     sess.Token,
		"url":       url,
	})
}

// handleCancelSession releases a session and its receiver early (e.g. when the
// user closes the extension overlay).
func (s *Server) handleCancelSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	token := r.URL.Query().Get("t")

	if _, err := s.store.Get(id, token); err != nil {
		httpx.WriteError(w, mapSessionError(err))
		return
	}

	s.hub.Unregister(id)
	s.store.Delete(id)
	w.WriteHeader(http.StatusNoContent)
}

// handleTransferWS upgrades the PC to a WebSocket receiver for a session.
func (s *Server) handleTransferWS(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("session")
	token := r.URL.Query().Get("token")

	if _, err := s.store.Get(id, token); err != nil {
		httpx.WriteError(w, mapSessionError(err))
		return
	}

	conn, err := s.hub.Upgrade(w, r, id)
	if err != nil {
		s.log.Warn("websocket upgrade failed", "error", err)
		return
	}

	go s.hub.HandleConn(r.Context(), id, conn, s.log)
}

// cryptoOverhead is the extra bytes added on top of the plaintext by the
// end-to-end encryption (24-byte nonce + 16-byte Poly1305 tag). The phone
// validates the plaintext against MaxFileSize, so the ciphertext accepted here
// must tolerate this overhead.
const cryptoOverhead = 40

// handleUpload receives the image from the phone and relays it to the PC.
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	token := r.URL.Query().Get("t")

	if _, err := s.store.Get(id, token); err != nil {
		httpx.WriteError(w, mapSessionError(err))
		return
	}

	// Bound the whole request body so an oversized upload fails early.
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxFileSize+1<<10)
	if err := r.ParseMultipartForm(s.cfg.MaxFileSize + 1<<10); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			httpx.WriteError(w, httpx.RequestEntityTooLarge("file_too_large", "file exceeds maximum allowed size"))
			return
		}
		httpx.WriteError(w, httpx.BadRequest("bad_request", "invalid multipart form"))
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("bad_request", "missing 'file' field"))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("bad_request", "failed to read file"))
		return
	}
	if int64(len(data)) > s.cfg.MaxFileSize+cryptoOverhead {
		httpx.WriteError(w, httpx.RequestEntityTooLarge("file_too_large", "file exceeds maximum allowed size"))
		return
	}

	name := header.Filename
	if v := r.FormValue("name"); v != "" {
		name = v
	}
	contentType := r.FormValue("contentType")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// The payload is end-to-end encrypted, so the server treats it as opaque
	// bytes and does not inspect or validate its contents. Only the file name
	// is normalized; type enforcement happens on the phone (before encrypting)
	// and on the PC (after decrypting).
	up := upload.Prepare(name, contentType, int64(len(data)))

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err = s.hub.Deliver(ctx, id, transfer.Meta{
		Name:        up.Name,
		ContentType: up.ContentType,
		Size:        up.Size,
	}, data)
	if err != nil {
		if errors.Is(err, transfer.ErrNoReceiver) {
			httpx.WriteError(w, httpx.Gone("no_receiver", "receiver is no longer connected"))
			return
		}
		s.log.Error("failed to deliver file", "error", err)
		httpx.WriteError(w, httpx.Internal("failed to deliver file"))
		return
	}

	// Delivery succeeded; the session is single-use and can be torn down.
	s.hub.Unregister(id)
	s.store.Delete(id)

	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleMobilePage serves the phone capture page after validating the token.
func (s *Server) handleMobilePage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	token := r.URL.Query().Get("t")

	if _, err := s.store.Get(id, token); err != nil {
		// Browsers expect HTML here, not JSON.
		http.Error(w, "Link expired or invalid. Scan the QR code again.", http.StatusGone)
		return
	}

	page, err := fs.ReadFile(s.web, "index.html")
	if err != nil {
		s.log.Error("failed to read mobile page", "error", err)
		httpx.WriteError(w, httpx.Internal("failed to load page"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}

// buildURL resolves the public base URL for QR codes. It prefers the
// configured BaseURL and otherwise derives one from the incoming request,
// honoring X-Forwarded-Proto behind proxies.
func (s *Server) buildURL(r *http.Request, path string) string {
	base := strings.TrimRight(s.cfg.BaseURL, "/")
	if base == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
		base = scheme + "://" + r.Host
	}
	return base + path
}

// mapSessionError converts session store errors into structured API errors.
func mapSessionError(err error) error {
	switch {
	case errors.Is(err, session.ErrExpired):
		return httpx.Gone("session_expired", "session expired")
	case errors.Is(err, session.ErrInvalidToken):
		return httpx.Unauthorized("invalid_token", "invalid token")
	case errors.Is(err, session.ErrNotFound):
		return httpx.NotFound("session_not_found", "session not found")
	default:
		return httpx.Internal("unexpected error")
	}
}
