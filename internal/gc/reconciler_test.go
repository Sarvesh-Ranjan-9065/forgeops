package gc

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/log"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/preview"
)

func TestCollectorSweep(t *testing.T) {
	ctx := context.Background()
	logger := log.New("error")

	now := time.Now()
	expiredTime := strconv.FormatInt(now.Add(-48*time.Hour).Unix(), 10)
	recentTime := strconv.FormatInt(now.Add(-1*time.Hour).Unix(), 10)

	clientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pr-1-old",
				Labels: map[string]string{
					preview.LabelPR:        "1",
					preview.LabelCreatedAt: expiredTime,
				},
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pr-2-new",
				Labels: map[string]string{
					preview.LabelPR:        "2",
					preview.LabelCreatedAt: recentTime,
				},
			},
		},
	)

	collector := &Collector{
		Clientset: clientset,
		Logger:    logger,
		TTL:       24 * time.Hour,
		Interval:  time.Hour,
	}

	collector.sweep(ctx)

	// pr-1-old was older than 24h, should be deleted
	_, err := clientset.CoreV1().Namespaces().Get(ctx, "pr-1-old", metav1.GetOptions{})
	require.Error(t, err)

	// pr-2-new was created 1h ago, should still exist
	ns, err := clientset.CoreV1().Namespaces().Get(ctx, "pr-2-new", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, "pr-2-new", ns.Name)
}
