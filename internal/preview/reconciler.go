package preview

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/go-github/v62/github"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/metrics"
)

// Reconciler converges preview namespaces with the set of open pull requests. It
// removes previews for closed PRs and re-enqueues provisioning for open PRs
// whose namespace is missing. Namespace labels are the only state.
type Reconciler struct {
	Clientset kubernetes.Interface
	GitHub    *github.Client
	Worker    *Worker
	Logger    *slog.Logger
	Owner     string
	Repo      string
	Interval  time.Duration
}

// Run reconciles immediately and then on every Interval tick until ctx is done.
func (r *Reconciler) Run(ctx context.Context) {
	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()
	r.reconcile(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.reconcile(ctx)
		}
	}
}

// reconcile tears down previews for closed PRs and re-enqueues missing ones.
func (r *Reconciler) reconcile(ctx context.Context) {
	openPRs, err := r.listOpenPRs(ctx)
	if err != nil {
		metrics.ReconcileErrorsTotal.WithLabelValues("reconcile").Inc()
		r.Logger.Error("reconcile list PRs", "error", err)
		return
	}
	nsList, err := r.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: LabelPR})
	if err != nil {
		metrics.ReconcileErrorsTotal.WithLabelValues("reconcile").Inc()
		r.Logger.Error("reconcile list namespaces", "error", err)
		return
	}

	live := map[int]bool{}
	for i := range nsList.Items {
		ns := &nsList.Items[i]
		prNum, err := strconv.Atoi(ns.Labels[LabelPR])
		if err != nil {
			continue
		}
		live[prNum] = true
		if _, ok := openPRs[prNum]; !ok {
			if err := Teardown(ctx, r.Logger, r.Clientset, prNum, ns.Labels[LabelService]); err != nil {
				r.Logger.Error("reconcile teardown", "pr", prNum, "error", err)
			}
		}
	}

	for num, e := range openPRs {
		if live[num] {
			continue
		}
		if !r.Worker.Enqueue(e) {
			r.Logger.Warn("reconcile enqueue dropped; queue full", "pr", num)
		}
	}
}

// listOpenPRs returns currently open pull requests keyed by number, expressed as
// provisioning events.
func (r *Reconciler) listOpenPRs(ctx context.Context) (map[int]Event, error) {
	out := map[int]Event{}
	opt := &github.PullRequestListOptions{
		State:       "open",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		prs, resp, err := r.GitHub.PullRequests.List(ctx, r.Owner, r.Repo, opt)
		if err != nil {
			return nil, err
		}
		for _, pr := range prs {
			out[pr.GetNumber()] = Event{
				Action:   "reopened",
				PRNumber: pr.GetNumber(),
				Service:  r.Repo,
				Owner:    r.Owner,
				Repo:     r.Repo,
				HeadRef:  pr.GetHead().GetRef(),
				CloneURL: pr.GetBase().GetRepo().GetCloneURL(),
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return out, nil
}
