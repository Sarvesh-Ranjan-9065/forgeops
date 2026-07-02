package gh

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/go-github/v62/github"
)

// CreateRepo creates a repository for the authenticated user. It is idempotent:
// an HTTP 422 response (repository already exists) is treated as success.
func CreateRepo(ctx context.Context, client *github.Client, name string, private bool) error {
	repo := &github.Repository{
		Name:     github.String(name),
		Private:  github.Bool(private),
		AutoInit: github.Bool(true),
	}
	// An empty org string creates the repo under the authenticated user.
	_, resp, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnprocessableEntity {
			return nil
		}
		var ge *github.ErrorResponse
		if errors.As(err, &ge) && ge.Response != nil &&
			ge.Response.StatusCode == http.StatusUnprocessableEntity {
			return nil
		}
		return err
	}
	return nil
}
