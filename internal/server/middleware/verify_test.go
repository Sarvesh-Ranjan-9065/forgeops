package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func computeSignature(secret, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	secret := []byte("test-secret-key")
	body := []byte(`{"action":"opened","number":42}`)

	t.Run("valid signature passes and body is preserved", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(signatureHeader, computeSignature(secret, body))

		called := false
		handler := VerifySignature(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			readBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, body, readBody)
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		require.True(t, called)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing signature rejected with 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))

		called := false
		handler := VerifySignature(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		handler.ServeHTTP(rec, req)
		require.False(t, called)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid signature rejected with 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(signatureHeader, "sha256=invalidhex0123456789abcdef0123456789abcdef0123456789abcdef0123456789")

		called := false
		handler := VerifySignature(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		handler.ServeHTTP(rec, req)
		require.False(t, called)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("wrong prefix rejected with 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(signatureHeader, "sha1=abcdef")

		called := false
		handler := VerifySignature(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		handler.ServeHTTP(rec, req)
		require.False(t, called)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
