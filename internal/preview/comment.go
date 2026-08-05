package preview

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v62/github"
)

// stickyMarker uniquely identifies the ForgeOps preview comment so it can be
// updated in place rather than duplicated.
const stickyMarker = "<!-- forgeops-preview -->"

// UpsertComment posts the preview URL as a single sticky PR comment, updating
// the existing comment if one is already present across all comment pages.
func UpsertComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int, url string) error {
	body := fmt.Sprintf("%s\nPreview environment is live: %s", stickyMarker, url)
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		comments, resp, err := client.Issues.ListComments(ctx, owner, repo, prNumber, opts)
		if err != nil {
			return fmt.Errorf("list comments: %w", err)
		}
		for _, c := range comments {
			if c.Body != nil && strings.Contains(*c.Body, stickyMarker) {
				_, _, err := client.Issues.EditComment(ctx, owner, repo, c.GetID(),
					&github.IssueComment{Body: github.String(body)})
				return err
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber,
		&github.IssueComment{Body: github.String(body)})
	return err
}
