package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/metrics"
	fwmw "github.com/sarvesh-ranjan-9065/forgeops/internal/server/middleware"
)

// RouterDeps holds the dependencies needed to build the HTTP router.
type RouterDeps struct {
	// WebhookSecret is the HMAC secret used to verify GitHub webhook signatures.
	WebhookSecret []byte
	// WebhookHandler processes verified webhook requests. If nil, a default
	// acknowledging handler is used (real processing is wired in Phase 4).
	WebhookHandler http.HandlerFunc
	// DashboardHandler serves the read-only dashboard UI and JSON API under
	// /dashboard. If nil, the dashboard is not mounted.
	DashboardHandler http.Handler
	// ReadinessCheck verifies downstream dependency health (e.g. Kubernetes API).
	// If nil or returns nil, /readyz reports 200 OK.
	ReadinessCheck func(context.Context) error
	// AdminUser and AdminPass optionally configure HTTP Basic Auth for
	// internal endpoints (/metrics and /dashboard).
	AdminUser string
	AdminPass string
	// Logger is used to log server events, such as readiness check errors.
	Logger *slog.Logger
}

// NewRouter builds the Chi router exposing health, readiness, webhook, and
// metrics endpoints. The webhook route is protected by HMAC verification.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		if deps.ReadinessCheck != nil {
			if err := deps.ReadinessCheck(req.Context()); err != nil {
				if deps.Logger != nil {
					deps.Logger.Warn("readiness check failed", "error", err)
				}
				http.Error(w, "not ready", http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	internalAuth := fwmw.ProtectInternal(deps.AdminUser, deps.AdminPass)
	r.Group(func(ir chi.Router) {
		ir.Use(internalAuth)
		ir.Handle("/metrics", promhttp.Handler())

		if deps.DashboardHandler != nil {
			ir.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/dashboard/", http.StatusMovedPermanently)
			})
			ir.Mount("/dashboard", deps.DashboardHandler)
		}
	})

	r.Group(func(pr chi.Router) {
		pr.Use(fwmw.VerifySignature(deps.WebhookSecret))
		handler := deps.WebhookHandler
		if handler == nil {
			handler = defaultWebhookHandler
		}
		pr.Post("/webhook", handler)
	})

	return r
}

// defaultWebhookHandler acknowledges a verified webhook. Event processing is
// wired in Phase 4.
func defaultWebhookHandler(w http.ResponseWriter, _ *http.Request) {
	metrics.WebhookEventsTotal.WithLabelValues("accepted").Inc()
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("accepted"))
}
