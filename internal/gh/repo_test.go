package gh

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/stretchr/testify/require"
)

func TestCreateRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("creates repo successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v3/user" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"login":"test-user"}`))
				return
			}
			if r.URL.Path == "/api/v3/user/repos" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"name":"new-repo"}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		client, err := github.NewClient(nil).WithEnterpriseURLs(server.URL+"/", server.URL+"/")
		require.NoError(t, err)

		err = CreateRepo(ctx, client, "test-user", "new-repo", false)
		require.NoError(t, err)
	})

	t.Run("idempotent on 422 already exists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v3/user" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"login":"test-user"}`))
				return
			}
			if r.URL.Path == "/api/v3/user/repos" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"message":"Repository creation failed.","errors":[{"resource":"Repository","code":"custom","field":"name","message":"name already exists on this account"}]}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		client, err := github.NewClient(nil).WithEnterpriseURLs(server.URL+"/", server.URL+"/")
		require.NoError(t, err)

		err = CreateRepo(ctx, client, "test-user", "existing-repo", false)
		require.NoError(t, err)
	})

	t.Run("returns error on other HTTP errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v3/user" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"login":"test-user"}`))
				return
			}
			if r.URL.Path == "/api/v3/user/repos" {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"message":"Internal server error"}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		client, err := github.NewClient(nil).WithEnterpriseURLs(server.URL+"/", server.URL+"/")
		require.NoError(t, err)

		err = CreateRepo(ctx, client, "test-user", "err-repo", false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "create repository")
	})
}
