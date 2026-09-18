package preview

import (
	"context"
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

		author := ""
		if user := pr.GetPullRequest().GetUser(); user != nil {
			author = user.GetLogin()
		}
		owner := pr.GetRepo().GetOwner().GetLogin()
		repo := pr.GetRepo().GetName()

		if !isAuthorAllowed(r.Context(), w.prov.GitHub, owner, repo, author) {
			logger.Warn("ignoring PR from untrusted author", "author", author, "repo", repo, "pr", pr.GetNumber())
			rw.WriteHeader(http.StatusOK)
			return
		}

		cloneURL := pr.GetRepo().GetCloneURL()
		if headRepo := pr.GetPullRequest().GetHead().GetRepo(); headRepo != nil && headRepo.GetCloneURL() != "" {
			cloneURL = headRepo.GetCloneURL()
		}

		e := Event{
			Action:   pr.GetAction(),
			PRNumber: pr.GetNumber(),
			Service:  repo,
			Owner:    owner,
			Repo:     repo,
			HeadRef:  pr.GetPullRequest().GetHead().GetRef(),
			CloneURL: cloneURL,
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

// isAuthorAllowed reports whether the author of a PR has permission to trigger
// a preview environment build (admin or write permission). An empty author is
// rejected immediately. If client is nil (such as in local unit tests), authorization
// is bypassed only when a non-empty author is provided.
func isAuthorAllowed(ctx context.Context, client *github.Client, owner, repo, author string) bool {
	if author == "" {
		return false
	}
	if client == nil {
		return true
	}
	perm, _, err := client.Repositories.GetPermissionLevel(ctx, owner, repo, author)
	if err != nil {
		return false
	}
	switch perm.GetPermission() {
	case "admin", "write":
		return true
	default:
		return false
	}
}
