package preview

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
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
	mu       sync.Mutex
	inflight map[int]bool
}

// NewWorker returns a Worker with a buffered queue of the given size.
func NewWorker(prov *Provisioner, buffer int) *Worker {
	return &Worker{
		prov:     prov,
		queue:    make(chan Event, buffer),
		logger:   prov.Logger,
		inflight: make(map[int]bool),
	}
}

// Enqueue adds an event to the work queue. It returns false if the queue is full
// so the caller can shed load instead of blocking the HTTP handler.
func (w *Worker) Enqueue(e Event) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.inflight[e.PRNumber] {
		w.logger.Debug("dropping duplicate; pr already queued", "pr", e.PRNumber)
		return true // existing queued work already covers this PR
	}
	select {
	case w.queue <- e:
		w.inflight[e.PRNumber] = true
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
			w.drain()
			return
		case e := <-w.queue:
			w.process(ctx, e)
		}
	}
}

// drain finishes any queued events using a bounded background context so a
// shutdown does not abandon already-accepted work.
func (w *Worker) drain() {
	drainCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for {
		select {
		case e := <-w.queue:
			w.process(drainCtx, e)
		default:
			return
		}
	}
}

// process handles a single event. Only PR-opening actions provision in Phase 4;
// closing/teardown is handled in Phase 5.
 func (w *Worker) process(ctx context.Context, e Event) {
	defer w.done(e.PRNumber)
	switch e.Action {
	case "closed":
		jobCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		if err := withRetry(jobCtx, w.logger, func(c context.Context) error {
			return Teardown(c, w.logger, w.prov.Clientset, e.PRNumber, e.Service)
		}); err != nil {
			metrics.WebhookEventsTotal.WithLabelValues("error").Inc()
			w.logger.Error("teardown failed", "pr", e.PRNumber, "error", err)
		}
	case "opened", "reopened", "synchronize":
		jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		err := withRetry(jobCtx, w.logger, func(c context.Context) error {
			return w.prov.Provision(c, e)
		})
		if err != nil {
			metrics.WebhookEventsTotal.WithLabelValues("error").Inc()
			w.logger.Error("provision failed", "pr", e.PRNumber, "error", err)
			if e.Action != "synchronize" {
				// Clean up partial resources so a failed first provision does not
				// linger until GC. A working env from a prior sync is left intact.
				if tErr := Teardown(jobCtx, w.logger, w.prov.Clientset, e.PRNumber, e.Service); tErr != nil {
					w.logger.Error("partial cleanup failed", "pr", e.PRNumber, "error", tErr)
				}
			}
		}
	default:
		w.logger.Debug("ignoring action", "action", e.Action)
	}
 }

// done clears the in-flight guard for a PR once its event finishes.
func (w *Worker) done(pr int) {
	w.mu.Lock()
	delete(w.inflight, pr)
	w.mu.Unlock()
}

// withRetry runs fn up to three times with capped exponential backoff, stopping
// early if ctx is canceled.
func withRetry(ctx context.Context, logger *slog.Logger, fn func(context.Context) error) error {
	const maxAttempts = 3
	const maxBackoff = 30 * time.Second
	backoff := 2 * time.Second
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = fn(ctx); err == nil {
			return nil
		}
		if attempt == maxAttempts {
			break
		}
		logger.Warn("operation failed; retrying", "attempt", attempt, "backoff", backoff.String(), "error", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
	return err
 }
