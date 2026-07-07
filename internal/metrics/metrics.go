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
