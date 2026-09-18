package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtectInternal(t *testing.T) {
	t.Run("basic auth configured: valid credentials pass", func(t *testing.T) {
		mw := ProtectInternal("admin", "secret")
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.SetBasicAuth("admin", "secret")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("basic auth configured: invalid credentials rejected with 401", func(t *testing.T) {
		mw := ProtectInternal("admin", "secret")
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.SetBasicAuth("admin", "wrongpassword")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Contains(t, rec.Header().Get("WWW-Authenticate"), "Basic")
	})

	t.Run("basic auth not configured: loopback/private allowed", func(t *testing.T) {
		mw := ProtectInternal("", "")
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Localhost / loopback
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		// Private network (RFC1918)
		req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.RemoteAddr = "10.0.0.5:12345"
		rec = httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("basic auth not configured: public IP rejected with 403", func(t *testing.T) {
		mw := ProtectInternal("", "")
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.RemoteAddr = "203.0.113.195:8080"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})
}
