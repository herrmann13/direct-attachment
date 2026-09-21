package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/direct-attachment-plugin/server/internal/api"
	"github.com/direct-attachment-plugin/server/internal/config"
	"github.com/direct-attachment-plugin/server/internal/session"
	"github.com/direct-attachment-plugin/server/internal/transfer"
	"github.com/direct-attachment-plugin/server/web"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	store := session.NewStore(cfg.SessionTTL)
	hub := transfer.NewHub()

	cleanupStop := make(chan struct{})
	go store.RunCleanup(30*time.Second, cleanupStop)
	defer close(cleanupStop)

	srv := api.New(cfg, store, hub, web.FS, log)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("server starting", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}

	log.Info("server stopped")
}
