package preview

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/v62/github"
	"k8s.io/client-go/kubernetes"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/metrics"
)

// Event describes a pull request event to be processed.
type Event struct {
	Action   string
	PRNumber int
	Service  string
	Owner    string
	Repo     string
	HeadRef  string
	CloneURL string
}

// Provisioner holds the dependencies for provisioning preview environments.
type Provisioner struct {
	Clientset kubernetes.Interface
	GitHub    *github.Client
	Logger    *slog.Logger
	Registry  string // e.g. "localhost:5000"
}

// Provision builds, deploys, and exposes a preview environment for the PR, then
// posts a sticky comment with the URL. Every step is idempotent.
func (p *Provisioner) Provision(ctx context.Context, e Event) error {
	ns, err := EnsureNamespace(ctx, p.Clientset, e.PRNumber, e.Service)
	if err != nil {
		return err
	}
	dir, err := Checkout(ctx, p.Logger, e.CloneURL, e.HeadRef)
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()

	image, err := BuildAndPush(ctx, p.Logger, p.Registry, e.Service, e.PRNumber, dir)
	if err != nil {
		return err
	}
	host := fmt.Sprintf("pr-%d-%s.127-0-0-1.nip.io", e.PRNumber, e.Service)
	if err := Deploy(ctx, p.Logger, ns, e.Service, filepath.Join(dir, "helm"), image, host); err != nil {
		return err
	}
	if err := WaitReady(ctx, p.Clientset, ns, e.Service); err != nil {
		return err
	}
	url := "http://" + host
	if err := UpsertComment(ctx, p.GitHub, e.Owner, e.Repo, e.PRNumber, url); err != nil {
		return err
	}
	p.Logger.Info("preview ready", "pr", e.PRNumber, "svc", e.Service, "url", url)
	return nil
}

// Worker serializes provisioning so the webhook handler can return immediately.
// Events are processed one at a time in FIFO order.
type Worker struct {
	prov   *Provisioner
	queue  chan Event
	logger *slog.Logger
}

// NewWorker returns a Worker with a buffered queue of the given size.
func NewWorker(prov *Provisioner, buffer int) *Worker {
	return &Worker{prov: prov, queue: make(chan Event, buffer), logger: prov.Logger}
}

// Enqueue adds an event to the work queue. It returns false if the queue is full
// so the caller can shed load instead of blocking the HTTP handler.
func (w *Worker) Enqueue(e Event) bool {
	select {
	case w.queue <- e:
		return true
	default:
		return false
	}
}

// Run processes queued events until ctx is canceled.
func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-w.queue:
			w.process(ctx, e)
		}
	}
}

// process handles a single event. Only PR-opening actions provision in Phase 4;
// closing/teardown is handled in Phase 5.
func (w *Worker) process(ctx context.Context, e Event) {
	switch e.Action {
	case "opened", "reopened", "synchronize":
		jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		if err := w.prov.Provision(jobCtx, e); err != nil {
			metrics.WebhookEventsTotal.WithLabelValues("error").Inc()
			w.logger.Error("provision failed", "pr", e.PRNumber, "error", err)
		}
	default:
		w.logger.Debug("ignoring action", "action", e.Action)
	}
}
