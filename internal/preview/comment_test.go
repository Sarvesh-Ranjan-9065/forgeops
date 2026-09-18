package preview

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/stretchr/testify/require"
)

func TestUpsertComment(t *testing.T) {
	ctx := context.Background()

	t.Run("creates comment when none exists", func(t *testing.T) {
		created := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/api/v3/repos/owner/repo/issues/1/comments" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
				return
			}
			if r.Method == http.MethodPost && r.URL.Path == "/api/v3/repos/owner/repo/issues/1/comments" {
				created = true
				var comment github.IssueComment
				_ = json.NewDecoder(r.Body).Decode(&comment)
				require.Contains(t, *comment.Body, stickyMarker)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":123}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		client, err := github.NewClient(nil).WithEnterpriseURLs(server.URL+"/", server.URL+"/")
		require.NoError(t, err)

		err = UpsertComment(ctx, client, "owner", "repo", 1, "http://preview.nip.io")
		require.NoError(t, err)
		require.True(t, created)
	})

	t.Run("edits existing sticky comment", func(t *testing.T) {
		edited := false
		existingID := int64(456)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/api/v3/repos/owner/repo/issues/2/comments" {
				w.Header().Set("Content-Type", "application/json")
				comments := []map[string]any{
					{"id": 111, "body": "regular comment"},
					{"id": existingID, "body": fmt.Sprintf("%s\nOld preview", stickyMarker)},
				}
				_ = json.NewEncoder(w).Encode(comments)
				return
			}
			if r.Method == http.MethodPatch && r.URL.Path == fmt.Sprintf("/api/v3/repos/owner/repo/issues/comments/%d", existingID) {
				edited = true
				var comment github.IssueComment
				_ = json.NewDecoder(r.Body).Decode(&comment)
				require.Contains(t, *comment.Body, "http://new-preview.nip.io")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":456}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		client, err := github.NewClient(nil).WithEnterpriseURLs(server.URL+"/", server.URL+"/")
		require.NoError(t, err)

		err = UpsertComment(ctx, client, "owner", "repo", 2, "http://new-preview.nip.io")
		require.NoError(t, err)
		require.True(t, edited)
	})
}

func TestDecommissionComment(t *testing.T) {
	ctx := context.Background()

	t.Run("nil client returns nil", func(t *testing.T) {
		err := DecommissionComment(ctx, nil, "owner", "repo", 1)
		require.NoError(t, err)
	})

	t.Run("updates sticky comment with decommission notice", func(t *testing.T) {
		decommissioned := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/api/v3/repos/owner/repo/issues/5/comments" {
				w.Header().Set("Content-Type", "application/json")
				comments := []map[string]any{
					{"id": 789, "body": fmt.Sprintf("%s\nActive", stickyMarker)},
				}
				_ = json.NewEncoder(w).Encode(comments)
				return
			}
			if r.Method == http.MethodPatch && r.URL.Path == "/api/v3/repos/owner/repo/issues/comments/789" {
				decommissioned = true
				var comment github.IssueComment
				_ = json.NewDecoder(r.Body).Decode(&comment)
				require.Contains(t, *comment.Body, "torn down")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":789}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		client, err := github.NewClient(nil).WithEnterpriseURLs(server.URL+"/", server.URL+"/")
		require.NoError(t, err)

		err = DecommissionComment(ctx, client, "owner", "repo", 5)
		require.NoError(t, err)
		require.True(t, decommissioned)
	})
}
