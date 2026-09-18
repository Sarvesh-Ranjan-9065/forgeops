package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/log"
)

func TestRouter(t *testing.T) {
	logger := log.New("error")
	secret := []byte("secret-key")

	deps := RouterDeps{
		WebhookSecret: secret,
		Logger:        logger,
	}
	r := NewRouter(deps)

	t.Run("healthz returns 200", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "ok", rec.Body.String())
	})

	t.Run("readyz returns 200 when check passes", func(t *testing.T) {
		readyDeps := deps
		readyDeps.ReadinessCheck = func(ctx context.Context) error {
			return nil
		}
		readyRouter := NewRouter(readyDeps)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		readyRouter.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "ready", rec.Body.String())
	})

	t.Run("readyz returns 503 and sanitized error when check fails", func(t *testing.T) {
		failingDeps := deps
		failingDeps.ReadinessCheck = func(ctx context.Context) error {
			return errors.New("k8s api server unreachable: https://10.96.0.1:443")
		}
		failingRouter := NewRouter(failingDeps)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		failingRouter.ServeHTTP(rec, req)
		require.Equal(t, http.StatusServiceUnavailable, rec.Code)
		require.Contains(t, rec.Body.String(), "not ready")
		require.NotContains(t, rec.Body.String(), "10.96.0.1")
	})

	t.Run("webhook rejects unauthenticated requests", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
