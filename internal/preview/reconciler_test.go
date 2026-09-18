package preview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/log"
)

func TestReconciler(t *testing.T) {
	ctx := context.Background()
	logger := log.New("error")

	t.Run("tears down namespaces for PRs no longer open", func(t *testing.T) {
		// Mock GitHub server returning no open PRs
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v3/repos/test-owner/test-repo/pulls" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		ghClient, err := github.NewClient(nil).WithEnterpriseURLs(server.URL+"/", server.URL+"/")
		require.NoError(t, err)

		clientset := fake.NewSimpleClientset(
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pr-99-test-repo",
					Labels: map[string]string{
						LabelPR:      "99",
						LabelService: "test-repo",
					},
				},
			},
		)

		worker := NewWorker(&Provisioner{Clientset: clientset}, 10)

		rec := &Reconciler{
			Clientset: clientset,
			GitHub:    ghClient,
			Worker:    worker,
			Logger:    logger,
			Owner:     "test-owner",
			Repo:      "test-repo",
		}

		rec.reconcile(ctx)

		// The namespace for closed PR 99 should have been deleted
		_, getErr := clientset.CoreV1().Namespaces().Get(ctx, "pr-99-test-repo", metav1.GetOptions{})
		require.Error(t, getErr)
	})
}
