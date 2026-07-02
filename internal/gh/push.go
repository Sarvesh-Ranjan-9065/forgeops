package gh

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
)

// PushTree creates a single commit containing every file in files and points
// refs/heads/main at it. Keys in files are repository-relative paths. The first
// push to an empty repository is handled by creating the commit with no base
// tree and creating the ref; subsequent pushes update the ref instead.
func PushTree(ctx context.Context, client *github.Client, owner, repo, message string, files map[string][]byte) error {
	entries := make([]*github.TreeEntry, 0, len(files))
	for path, content := range files {
		blob, _, err := client.Git.CreateBlob(ctx, owner, repo, &github.Blob{
			Content:  github.String(string(content)),
			Encoding: github.String("utf-8"),
		})
		if err != nil {
			return fmt.Errorf("create blob %s: %w", path, err)
		}
		entries = append(entries, &github.TreeEntry{
			Path: github.String(path),
			Mode: github.String("100644"),
			Type: github.String("blob"),
			SHA:  blob.SHA,
		})
	}

	// Empty base tree string is correct for the first commit in a new repo.
	fmt.Printf("DEBUG: tree has %d entries, first entry: %+v\n", len(entries), entries[0])
	tree, _, err := client.Git.CreateTree(ctx, owner, repo, "", entries)
	if err != nil {
		return fmt.Errorf("create tree: %w", err)
	}

	commit, _, err := client.Git.CreateCommit(ctx, owner, repo, &github.Commit{
		Message: github.String(message),
		Tree:    tree,
	}, nil)
	if err != nil {
		return fmt.Errorf("create commit: %w", err)
	}

	newRef := &github.Reference{
		Ref:    github.String("refs/heads/main"),
		Object: &github.GitObject{SHA: commit.SHA},
	}
	if _, _, err := client.Git.CreateRef(ctx, owner, repo, newRef); err != nil {
		// The ref already exists (re-push); force-update it instead.
		if _, _, uerr := client.Git.UpdateRef(ctx, owner, repo, newRef, true); uerr != nil {
			return fmt.Errorf("create/update ref: %w", uerr)
		}
	}
	return nil
}
