// Package server provides the ForgeOps HTTP server with graceful shutdown and
// the route table for the webhook service.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Server wraps an http.Server with context-aware graceful shutdown.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// New returns a Server listening on addr that serves handler.
func New(addr string, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Run starts the server and blocks until ctx is canceled, then shuts down
// gracefully within a bounded timeout.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server listening", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		s.logger.Info("server shutting down")
		return s.httpServer.Shutdown(shutdownCtx)
	}
}
