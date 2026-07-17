// Package preview provisions and manages per-PR preview environments.
package preview

import (
	"context"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Labels applied to every preview namespace. They are the source of truth for
// reconciliation and TTL garbage collection; there is no database.
const (
	LabelPR        = "forgeops.io/pr"
	LabelService   = "forgeops.io/svc"
	LabelCreatedAt = "forgeops.io/created-at"
)

// NamespaceName returns the deterministic namespace name for a PR and service.
func NamespaceName(prNumber int, svc string) string {
	return fmt.Sprintf("pr-%d-%s", prNumber, svc)
}

// EnsureNamespace idempotently creates the preview namespace for a PR and
// returns its name. An AlreadyExists error is treated as success. The created-at
// label stores Unix seconds (digits only) because label values cannot contain
// the colons of an RFC3339 timestamp.
func EnsureNamespace(ctx context.Context, cs kubernetes.Interface, prNumber int, svc string) (string, error) {
	name := NamespaceName(prNumber, svc)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				LabelPR:        strconv.Itoa(prNumber),
				LabelService:   svc,
				LabelCreatedAt: strconv.FormatInt(time.Now().Unix(), 10),
			},
		},
	}
	if _, err := cs.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil &&
		!apierrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("create namespace %s: %w", name, err)
	}
	return name, nil
}
