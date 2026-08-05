package preview

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentAvailable(t *testing.T) {
	t.Run("deployment available condition true", func(t *testing.T) {
		dep := &appsv1.Deployment{
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{
						Type:   appsv1.DeploymentAvailable,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
		require.True(t, deploymentAvailable(dep))
	})

	t.Run("deployment available condition false", func(t *testing.T) {
		dep := &appsv1.Deployment{
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{
						Type:   appsv1.DeploymentAvailable,
						Status: corev1.ConditionFalse,
					},
				},
			},
		}
		require.False(t, deploymentAvailable(dep))
	})
}

func TestWaitReadyTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	clientset := fake.NewSimpleClientset()
	err := WaitReady(ctx, clientset, "test-ns", "test-dep")
	require.Error(t, err)
}
