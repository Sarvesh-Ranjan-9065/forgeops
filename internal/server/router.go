package server

import (
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
	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	r.Handle("/metrics", promhttp.Handler())

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
