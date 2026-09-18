package gh

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/stretchr/testify/require"
)

func TestPushTree(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v3/repos/test-owner/test-repo/git/blobs":
			_, _ = w.Write([]byte(`{"sha":"blob-sha-123"}`))
		case "/api/v3/repos/test-owner/test-repo/git/trees":
			_, _ = w.Write([]byte(`{"sha":"tree-sha-456"}`))
		case "/api/v3/repos/test-owner/test-repo/git/commits":
			_, _ = w.Write([]byte(`{"sha":"commit-sha-789"}`))
		case "/api/v3/repos/test-owner/test-repo/git/refs":
			_, _ = w.Write([]byte(`{"ref":"refs/heads/main","object":{"sha":"commit-sha-789"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := github.NewClient(nil).WithEnterpriseURLs(server.URL+"/", server.URL+"/")
	require.NoError(t, err)

	files := map[string][]byte{
		"README.md": []byte("# Test Repo"),
		"main.go":   []byte("package main"),
	}

	err = PushTree(ctx, client, "test-owner", "test-repo", "initial commit", files)
	require.NoError(t, err)
}
