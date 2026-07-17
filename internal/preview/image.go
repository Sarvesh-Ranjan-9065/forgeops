package preview

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// ImageRef returns the fully-qualified image reference for a PR build.
func ImageRef(registry, svc string, prNumber int) string {
	return fmt.Sprintf("%s/%s:pr-%d", registry, svc, prNumber)
}

// Checkout shallow-clones cloneURL at ref into a new temp directory and returns
// the directory path. The caller is responsible for removing it.
func Checkout(ctx context.Context, logger *slog.Logger, cloneURL, ref string) (string, error) {
	dir, err := os.MkdirTemp("", "forgeops-build-")
	if err != nil {
		return "", err
	}
	if err := run(ctx, logger, "git", "clone", "--depth", "1", "--branch", ref, cloneURL, dir); err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("clone %s@%s: %w", cloneURL, ref, err)
	}
	return dir, nil
}

// BuildAndPush builds the image from contextDir and pushes it to the registry.
// It returns the pushed image reference.
func BuildAndPush(ctx context.Context, logger *slog.Logger, registry, svc string, prNumber int, contextDir string) (string, error) {
	image := ImageRef(registry, svc, prNumber)
	if err := run(ctx, logger, "docker", "build", "-t", image, contextDir); err != nil {
		return "", fmt.Errorf("docker build: %w", err)
	}
	if err := run(ctx, logger, "docker", "push", image); err != nil {
		return "", fmt.Errorf("docker push: %w", err)
	}
	return image, nil
}

// run executes a command, capturing combined output for diagnostics.
func run(ctx context.Context, logger *slog.Logger, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("command failed", "cmd", name, "args", args, "output", string(out))
		return err
	}
	logger.Debug("command ok", "cmd", name, "args", args)
	return nil
}
