package preview

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/server/middleware"
)

func TestWebhookHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	secret := []byte("devsecret")

	t.Run("valid signature and PR event enqueued", func(t *testing.T) {
		prov := &Provisioner{Logger: logger}
		worker := NewWorker(prov, 10)
		handler := NewWebhookHandler(worker, logger)
		wrapped := middleware.VerifySignature(secret)(handler)

		payload := `{
			"action": "opened",
			"number": 1,
			"pull_request": {
				"number": 1,
				"head": {
					"ref": "test-pr"
				}
			},
			"repository": {
				"name": "demo",
				"owner": {
					"login": "Sarvesh-Ranjan-9065"
				},
				"clone_url": "https://github.com/Sarvesh-Ranjan-9065/demo.git"
			}
		}`

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(payload)))
		req.Header.Set("X-GitHub-Event", "pull_request")

		mac := hmac.New(sha256.New, secret)
		mac.Write([]byte(payload))
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Hub-Signature-256", signature)

		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)

		select {
		case e := <-worker.queue:
			require.Equal(t, "opened", e.Action)
			require.Equal(t, 1, e.PRNumber)
			require.Equal(t, "demo", e.Service)
		default:
			t.Fatal("expected event in worker queue")
		}
	})

	t.Run("invalid signature rejected with 401", func(t *testing.T) {
		prov := &Provisioner{Logger: logger}
		worker := NewWorker(prov, 10)
		handler := NewWebhookHandler(worker, logger)
		wrapped := middleware.VerifySignature(secret)(handler)

		payload := `{"action": "opened"}`
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(payload)))
		req.Header.Set("X-GitHub-Event", "pull_request")
		req.Header.Set("X-Hub-Signature-256", "sha256=invalid")

		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("non-PR event ignored with 204", func(t *testing.T) {
		prov := &Provisioner{Logger: logger}
		worker := NewWorker(prov, 10)
		handler := NewWebhookHandler(worker, logger)
		wrapped := middleware.VerifySignature(secret)(handler)

		payload := `{"action": "completed"}`
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(payload)))
		req.Header.Set("X-GitHub-Event", "workflow_job")

		mac := hmac.New(sha256.New, secret)
		mac.Write([]byte(payload))
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Hub-Signature-256", signature)

		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)
	})
}
