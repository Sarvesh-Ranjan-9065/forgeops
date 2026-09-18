package dashboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestDashboardRoutes(t *testing.T) {
	logger := log.New("error")
	clientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pr-10-service",
				Labels: map[string]string{
					preview.LabelPR:        "10",
					preview.LabelService:   "service",
					preview.LabelCreatedAt: strconv.FormatInt(time.Now().Unix(), 10),
				},
			},
		},
	)

	handler := &Handler{
		Clientset: clientset,
		Logger:    logger,
		TTL:       24 * time.Hour,
	}
	router := handler.Routes()

	t.Run("serves dashboard index html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Header().Get("Content-Type"), "text/html")
		require.Contains(t, rec.Body.String(), "<!DOCTYPE html>")
	})

	t.Run("serves api environments json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/environments", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp APIResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Environments, 1)
		require.Equal(t, 10, resp.Environments[0].PR)
		require.Equal(t, "service", resp.Environments[0].Service)
	})
}
