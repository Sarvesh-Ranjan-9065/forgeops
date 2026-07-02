// Package gh wraps the GitHub API client and the repository operations used by
// ForgeOps.
package gh

import (
	"context"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// NewClient returns an OAuth2-authenticated GitHub client.
func NewClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return github.NewClient(oauth2.NewClient(ctx, ts))
}
