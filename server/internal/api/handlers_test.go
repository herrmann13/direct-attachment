package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/direct-attachment-plugin/server/internal/config"
	"github.com/direct-attachment-plugin/server/internal/session"
	"github.com/direct-attachment-plugin/server/internal/transfer"
	"github.com/gorilla/websocket"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	cfg := config.Config{
		Port:        "8080",
		SessionTTL:  time.Minute,
		MaxFileSize: 15 << 20,
	}
	store := session.NewStore(cfg.SessionTTL)
	hub := transfer.NewHub()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	webFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html></html>")},
		"style.css":  &fstest.MapFile{Data: []byte("body{}")},
		"app.js":     &fstest.MapFile{Data: []byte("")},
	}

	srv := New(cfg, store, hub, webFS, log)
	ts := httptest.NewServer(srv.Routes())
	t.Cleanup(ts.Close)
	return ts
}

func createSession(t *testing.T, ts *httptest.Server) (id, token string) {
	t.Helper()
	res, err := http.Post(ts.URL+"/api/sessions", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/sessions: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/sessions status = %d, want 201", res.StatusCode)
	}

	var body struct {
		SessionID string `json:"sessionId"`
		Token     string `json:"token"`
		URL       string `json:"url"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if body.SessionID == "" || body.Token == "" || body.URL == "" {
		t.Fatalf("incomplete session response: %+v", body)
	}
	return body.SessionID, body.Token
}

func dialWS(t *testing.T, ts *httptest.Server, id, token string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/transfer?session=" + id + "&token=" + token
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func uploadFile(t *testing.T, ts *httptest.Server, id, token, name, contentType string, data []byte) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	if err := mw.WriteField("contentType", contentType); err != nil {
		t.Fatalf("write contentType: %v", err)
	}

	fw, err := mw.CreateFormFile("file", name)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	_ = mw.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/upload/"+id+"?t="+token, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/upload: %v", err)
	}
	return res
}

// readMeta reads the metadata frame and returns the parsed fields.
func readMeta(t *testing.T, conn *websocket.Conn) (name, contentType string) {
	t.Helper()
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	var meta struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		ContentType string `json:"contentType"`
		Size        int64  `json:"size"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("unmarshal meta: %v", err)
	}
	if meta.Type != "meta" {
		t.Fatalf("meta type = %q, want meta", meta.Type)
	}
	return meta.Name, meta.ContentType
}

func TestUploadRelaysOpaqueBytes(t *testing.T) {
	ts := newTestServer(t)
	id, token := createSession(t, ts)
	conn := dialWS(t, ts, id, token)

	// The payload is encrypted on the client; the server must relay it
	// byte-for-byte without inspecting its contents.
	data := []byte("opaque-encrypted-bytes-that-are-not-an-image")
	res := uploadFile(t, ts, id, token, "foto.png", "image/png", data)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", res.StatusCode)
	}

	name, contentType := readMeta(t, conn)
	if name != "foto.png" || contentType != "image/png" {
		t.Fatalf("unexpected meta: name=%q contentType=%q", name, contentType)
	}

	mt, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if mt != websocket.BinaryMessage {
		t.Fatalf("message type = %d, want binary (%d)", mt, websocket.BinaryMessage)
	}
	if !bytes.Equal(payload, data) {
		t.Fatal("payload mismatch")
	}

	_, doneRaw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read done: %v", err)
	}
	if !strings.Contains(string(doneRaw), "done") {
		t.Fatalf("done frame = %q", string(doneRaw))
	}
}

func TestUploadAddsExtension(t *testing.T) {
	ts := newTestServer(t)
	id, token := createSession(t, ts)
	conn := dialWS(t, ts, id, token)

	res := uploadFile(t, ts, id, token, "foto", "image/jpeg", []byte("x"))
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", res.StatusCode)
	}

	name, _ := readMeta(t, conn)
	if name != "foto.jpg" {
		t.Fatalf("name = %q, want foto.jpg", name)
	}
}

func TestUploadInvalidToken(t *testing.T) {
	ts := newTestServer(t)
	id, _ := createSession(t, ts)

	res := uploadFile(t, ts, id, "wrong-token", "foto.png", "image/png", []byte("x"))
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.StatusCode)
	}
}

func TestUploadNoReceiver(t *testing.T) {
	ts := newTestServer(t)
	id, token := createSession(t, ts)

	// No WebSocket connected, so delivery must fail with 410.
	res := uploadFile(t, ts, id, token, "foto.png", "image/png", []byte("x"))
	defer res.Body.Close()
	if res.StatusCode != http.StatusGone {
		t.Fatalf("status = %d, want 410", res.StatusCode)
	}
}

func TestCORS(t *testing.T) {
	ts := newTestServer(t)

	// Simple cross-origin request must carry the CORS allow header.
	res, err := http.Post(ts.URL+"/api/sessions", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	res.Body.Close()
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}

	// Preflighted request (DELETE) must be answered with 204 and the method.
	req, _ := http.NewRequest(http.MethodOptions, ts.URL+"/api/sessions/xyz", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "DELETE")
	pre, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS: %v", err)
	}
	pre.Body.Close()
	if pre.StatusCode != http.StatusNoContent {
		t.Fatalf("OPTIONS status = %d, want 204", pre.StatusCode)
	}
	if got := pre.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, "DELETE") {
		t.Fatalf("Access-Control-Allow-Methods = %q, want to contain DELETE", got)
	}
}

func TestCancelSession(t *testing.T) {
	ts := newTestServer(t)
	id, token := createSession(t, ts)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/sessions/"+id+"?t="+token, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want 204", res.StatusCode)
	}

	// Uploading after cancellation must be rejected.
	up := uploadFile(t, ts, id, token, "foto.png", "image/png", []byte("x"))
	defer up.Body.Close()
	if up.StatusCode != http.StatusUnauthorized && up.StatusCode != http.StatusNotFound {
		t.Fatalf("upload after cancel status = %d, want 401/404", up.StatusCode)
	}
}
