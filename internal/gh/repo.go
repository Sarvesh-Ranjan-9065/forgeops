package gh

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v62/github"
)

// CreateRepo creates a repository under owner (an organization or the authenticated
// user). It is idempotent: an HTTP 422 response (repository already exists) is treated
// as success.
func CreateRepo(ctx context.Context, client *github.Client, owner, name string, private bool) error {
	repo := &github.Repository{
		Name:     github.String(name),
		Private:  github.Bool(private),
		AutoInit: github.Bool(true),
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("get authenticated user: %w", err)
	}

	org := ""
	if !strings.EqualFold(user.GetLogin(), owner) {
		org = owner
	}

	_, resp, err := client.Repositories.Create(ctx, org, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnprocessableEntity {
			return nil
		}
		var ge *github.ErrorResponse
		if errors.As(err, &ge) && ge.Response != nil &&
			ge.Response.StatusCode == http.StatusUnprocessableEntity {
			return nil
		}
		return fmt.Errorf("create repository %s/%s: %w", owner, name, err)
	}
	return nil
}
