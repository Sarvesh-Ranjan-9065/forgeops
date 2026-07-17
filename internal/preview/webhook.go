package preview

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/google/go-github/v62/github"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/metrics"
)

// NewWebhookHandler returns an http.HandlerFunc that parses GitHub pull request
// events and enqueues provisioning work. It responds quickly (202) so GitHub
// does not time out; the actual work happens on the Worker goroutine.
func NewWebhookHandler(w *Worker, logger *slog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, "cannot read body", http.StatusBadRequest)
			return
		}
		raw, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			http.Error(rw, "cannot parse webhook", http.StatusBadRequest)
			return
		}
		pr, ok := raw.(*github.PullRequestEvent)
		if !ok {
			rw.WriteHeader(http.StatusNoContent) // not a PR event; ignore
			return
		}
		e := Event{
			Action:   pr.GetAction(),
			PRNumber: pr.GetNumber(),
			Service:  pr.GetRepo().GetName(),
			Owner:    pr.GetRepo().GetOwner().GetLogin(),
			Repo:     pr.GetRepo().GetName(),
			HeadRef:  pr.GetPullRequest().GetHead().GetRef(),
			CloneURL: pr.GetRepo().GetCloneURL(),
		}
		if !w.Enqueue(e) {
			metrics.WebhookEventsTotal.WithLabelValues("error").Inc()
			http.Error(rw, "queue full", http.StatusServiceUnavailable)
			return
		}
		metrics.WebhookEventsTotal.WithLabelValues("accepted").Inc()
		logger.Info("enqueued pr event", "action", e.Action, "pr", e.PRNumber, "svc", e.Service)
		rw.WriteHeader(http.StatusAccepted)
	}
}
