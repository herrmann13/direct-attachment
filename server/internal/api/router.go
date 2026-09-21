// Package api wires the HTTP surface: routing, handlers, and middleware.
package api

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/direct-attachment-plugin/server/internal/config"
	"github.com/direct-attachment-plugin/server/internal/session"
	"github.com/direct-attachment-plugin/server/internal/transfer"
)

// Server holds the dependencies shared by every handler.
type Server struct {
	cfg   config.Config
	store *session.Store
	hub   *transfer.Hub
	web   fs.FS
	log   *slog.Logger
}

// New builds a Server with the given dependencies.
func New(cfg config.Config, store *session.Store, hub *transfer.Hub, web fs.FS, log *slog.Logger) *Server {
	return &Server{
		cfg:   cfg,
		store: store,
		hub:   hub,
		web:   web,
		log:   log,
	}
}

// Routes assembles the HTTP handler with all routes and middleware.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/sessions", s.handleCreateSession)
	mux.HandleFunc("DELETE /api/sessions/{id}", s.handleCancelSession)
	mux.HandleFunc("POST /api/upload/{id}", s.handleUpload)
	mux.HandleFunc("GET /ws/transfer", s.handleTransferWS)
	mux.HandleFunc("GET /s/{id}", s.handleMobilePage)

	// Static assets for the mobile page (style.css, app.js).
	mux.Handle("GET /", http.FileServer(http.FS(s.web)))

	return s.withCORS(s.withRecovery(s.withLogging(mux)))
}
