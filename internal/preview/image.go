package preview

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// ImageRef returns the fully-qualified image reference for a PR build.
func ImageRef(registry, svc string, prNumber int) string {
	return fmt.Sprintf("%s/%s:pr-%d", registry, svc, prNumber)
}

// Checkout shallow-clones cloneURL at ref into a new temp directory and returns
// the directory path. The caller is responsible for removing it. If token is non-empty
// and cloneURL is an HTTPS URL, the token is passed via Git HTTP authorization.
func Checkout(ctx context.Context, logger *slog.Logger, token, cloneURL, ref string) (string, error) {
	dir, err := os.MkdirTemp("", "forgeops-build-")
	if err != nil {
		return "", fmt.Errorf("create temp build dir: %w", err)
	}

	args := []string{"clone", "--depth", "1", "--branch", ref}
	if token != "" && strings.HasPrefix(cloneURL, "https://") {
		auth := base64.StdEncoding.EncodeToString([]byte("x-access-token:" + token))
		args = append([]string{"-c", fmt.Sprintf("http.extraheader=AUTHORIZATION: basic %s", auth)}, args...)
	}
	args = append(args, "--", cloneURL, dir)

	if err := run(ctx, logger, "git", args...); err != nil {
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
		logger.Error("command failed", "cmd", name, "args", redactArgs(args), "output", string(out))
		return fmt.Errorf("execute %s: %w", name, err)
	}
	logger.Debug("command ok", "cmd", name, "args", redactArgs(args))
	return nil
}

// redactArgs masks sensitive tokens or authorization headers in command arguments.
func redactArgs(args []string) []string {
	safe := make([]string, len(args))
	for i, arg := range args {
		if strings.Contains(arg, "AUTHORIZATION:") || strings.Contains(arg, "extraheader=") {
			safe[i] = "[REDACTED]"
		} else {
			safe[i] = arg
		}
	}
	return safe
}
