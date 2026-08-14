// Package metrics defines and registers the Prometheus metrics exposed by the
// ForgeOps webhook server. Metrics register with the default registry, which
// the /metrics endpoint serves.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// WebhookEventsTotal counts received webhook events, labeled by result
// ("accepted", "rejected", or "error").
var WebhookEventsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "forgeops_webhook_events_total",
		Help: "Total number of webhook events received, labeled by result.",
	},
	[]string{"result"},
)

// GCDeletionsTotal counts preview namespaces deleted by the TTL collector.
var GCDeletionsTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "forgeops_gc_deletions_total",
		Help: "Total number of expired preview namespaces deleted by GC.",
	},
)

// EnvsActive tracks the number of preview environments currently active.
var EnvsActive = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "forgeops_envs_active",
		Help: "Number of preview environments currently active.",
	},
)

// ProvisionDuration measures end-to-end preview provision time in seconds.
var ProvisionDuration = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "forgeops_provision_duration_seconds",
		Help:    "End-to-end preview provision duration in seconds.",
		Buckets: prometheus.ExponentialBuckets(5, 2, 7), // ~5s .. ~320s
	},
)

// ReconcileErrorsTotal counts reconcile and GC errors, labeled by component.
var ReconcileErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "forgeops_reconcile_errors_total",
		Help: "Total reconcile and GC errors, labeled by component.",
	},
	[]string{"component"},
)
