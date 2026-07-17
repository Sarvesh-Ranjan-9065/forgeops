package preview

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WaitReady blocks until the named Deployment reports Available or ctx is done.
// It polls because helm --wait already streams progress; this is the explicit
// secondary guard the control plane relies on.
func WaitReady(ctx context.Context, cs kubernetes.Interface, namespace, name string) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		dep, err := cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err == nil && deploymentAvailable(dep) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for deployment %s/%s: %w", namespace, name, ctx.Err())
		case <-ticker.C:
		}
	}
}

// deploymentAvailable reports whether a Deployment has the Available condition.
func deploymentAvailable(dep *appsv1.Deployment) bool {
	for _, c := range dep.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
