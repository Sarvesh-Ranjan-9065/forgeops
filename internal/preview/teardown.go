package preview

import (
	"context"
	"fmt"
	"log/slog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Teardown removes the preview environment for a PR: it uninstalls the Helm
// release and deletes the namespace. Both steps tolerate "not found" so repeated
// teardowns are safe.
func Teardown(ctx context.Context, logger *slog.Logger, cs kubernetes.Interface, prNumber int, svc string) error {
	ns := NamespaceName(prNumber, svc)

	// helm uninstall is best-effort; deleting the namespace removes everything
	// regardless, so a missing release is not fatal.
	if err := run(ctx, logger, "helm", "uninstall", svc, "--namespace", ns, "--ignore-not-found"); err != nil {
		logger.Warn("helm uninstall failed; continuing to namespace deletion", "ns", ns, "error", err)
	}

	if err := cs.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{}); err != nil &&
		!apierrors.IsNotFound(err) {
		return fmt.Errorf("delete namespace %s: %w", ns, err)
	}
	logger.Info("preview torn down", "pr", prNumber, "svc", svc, "ns", ns)
	return nil
}
