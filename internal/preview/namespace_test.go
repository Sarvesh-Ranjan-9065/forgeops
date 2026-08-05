package preview

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNamespaceName(t *testing.T) {
	name := NamespaceName(42, "frontend")
	require.Equal(t, "pr-42-frontend", name)
}

func TestEnsureNamespace(t *testing.T) {
	ctx := context.Background()
	clientset := fake.NewSimpleClientset()

	name, err := EnsureNamespace(ctx, clientset, 101, "api-service")
	require.NoError(t, err)
	require.Equal(t, "pr-101-api-service", name)

	// Idempotency check: calling a second time on an existing namespace
	name2, err := EnsureNamespace(ctx, clientset, 101, "api-service")
	require.NoError(t, err)
	require.Equal(t, name, name2)

	// Verify required namespace labels
	ns, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, "101", ns.Labels[LabelPR])
	require.Equal(t, "api-service", ns.Labels[LabelService])
	require.NotEmpty(t, ns.Labels[LabelCreatedAt])
}
