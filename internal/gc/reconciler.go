// Package gc garbage-collects expired preview environments based on their
// creation-time label. It is stateless: namespace labels are the only state.
package gc

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/metrics"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/preview"
)

// Collector periodically deletes preview namespaces older than TTL.
type Collector struct {
	Clientset kubernetes.Interface
	Logger    *slog.Logger
	TTL       time.Duration
	Interval  time.Duration
}

// Run executes a sweep immediately and then on every Interval tick until ctx is
// canceled.
func (c *Collector) Run(ctx context.Context) {
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()
	c.sweep(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.sweep(ctx)
		}
	}
}

// sweep deletes every preview namespace whose age exceeds the TTL.
func (c *Collector) sweep(ctx context.Context) {
	list, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: preview.LabelPR,
	})
	if err != nil {
		c.Logger.Error("gc list namespaces", "error", err)
		return
	}
	cutoff := time.Now().Add(-c.TTL)
	for i := range list.Items {
		ns := &list.Items[i]
		created, ok := createdAt(ns)
		if !ok || created.After(cutoff) {
			continue
		}
		if err := c.delete(ctx, ns.Name); err != nil {
			c.Logger.Error("gc delete namespace", "ns", ns.Name, "error", err)
			continue
		}
		metrics.GCDeletionsTotal.Inc()
		c.Logger.Info("gc deleted expired preview", "ns", ns.Name, "age", time.Since(created).String())
	}
}

// delete removes a namespace, tolerating an already-deleted namespace.
func (c *Collector) delete(ctx context.Context, name string) error {
	if err := c.Clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{}); err != nil &&
		!apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// createdAt parses the Unix-seconds creation label written at provision time.
func createdAt(ns *corev1.Namespace) (time.Time, bool) {
	raw, ok := ns.Labels[preview.LabelCreatedAt]
	if !ok {
		return time.Time{}, false
	}
	secs, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(secs, 0), true
}