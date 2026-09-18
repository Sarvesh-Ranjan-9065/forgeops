// Package preview provisions and manages per-PR preview environments.
package preview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
// If the combined name exceeds Kubernetes' 63-character limit, the service name
// is deterministically truncated and suffixed with an 8-character hash.
func NamespaceName(prNumber int, svc string) string {
	base := fmt.Sprintf("pr-%d-%s", prNumber, svc)
	if len(base) <= 63 {
		return base
	}
	h := sha256.Sum256([]byte(svc))
	suffix := "-" + hex.EncodeToString(h[:])[:8]
	prefix := fmt.Sprintf("pr-%d-", prNumber)
	maxSvcLen := 63 - len(prefix) - len(suffix)
	if maxSvcLen < 1 {
		return (prefix + hex.EncodeToString(h[:]))[:63]
	}
	trimmedSvc := strings.TrimRight(svc[:maxSvcLen], "-")
	return fmt.Sprintf("%s%s%s", prefix, trimmedSvc, suffix)
}

// sanitizeLabel truncates a string to at most 63 characters, appending a deterministic
// hash suffix if truncation was necessary to satisfy Kubernetes label value limits.
func sanitizeLabel(val string) string {
	if len(val) <= 63 {
		return val
	}
	h := sha256.Sum256([]byte(val))
	suffix := "-" + hex.EncodeToString(h[:])[:8]
	return strings.TrimRight(val[:63-len(suffix)], "-") + suffix
}

// EnsureNamespace idempotently creates the preview namespace for a PR, applies
// a protective ResourceQuota to prevent cluster resource starvation, and returns
// the namespace name. An AlreadyExists error is treated as success. The created-at
// label stores Unix seconds (digits only) because label values cannot contain
// the colons of an RFC3339 timestamp.
func EnsureNamespace(ctx context.Context, cs kubernetes.Interface, prNumber int, svc string) (string, error) {
	name := NamespaceName(prNumber, svc)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				LabelPR:        strconv.Itoa(prNumber),
				LabelService:   sanitizeLabel(svc),
				LabelCreatedAt: strconv.FormatInt(time.Now().Unix(), 10),
			},
		},
	}
	if _, err := cs.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil &&
		!apierrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("create namespace %s: %w", name, err)
	}

	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "forgeops-quota",
			Namespace: name,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourcePods:   resource.MustParse("10"),
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		},
	}
	if _, err := cs.CoreV1().ResourceQuotas(name).Create(ctx, quota, metav1.CreateOptions{}); err != nil &&
		!apierrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("create resource quota in %s: %w", name, err)
	}

	return name, nil
}
