package preview

import (
	"context"
	"log/slog"
	"strings"
)

// Deploy installs or upgrades the per-PR Helm release into namespace, overriding
// the image and ingress host with PR-specific values.
func Deploy(ctx context.Context, logger *slog.Logger, namespace, svc, chartPath, image, host string) error {
	args := []string{
		"upgrade", "--install", svc, chartPath,
		"--namespace", namespace,
		"--create-namespace",
		"--set", "image.repository=" + imageRepo(image),
		"--set", "image.tag=" + imageTag(image),
		"--set", "ingress.host=" + host,
		"--wait", "--timeout", "120s",
	}
	return run(ctx, logger, "helm", args...)
}

// imageRepo returns the repository portion of a registry/name:tag reference.
func imageRepo(image string) string {
	lastColon := strings.LastIndex(image, ":")
	lastSlash := strings.LastIndex(image, "/")
	if lastColon > lastSlash {
		return image[:lastColon]
	}
	return image
}

// imageTag returns the tag portion of a registry/name:tag reference.
func imageTag(image string) string {
	lastColon := strings.LastIndex(image, ":")
	lastSlash := strings.LastIndex(image, "/")
	if lastColon > lastSlash {
		return image[lastColon+1:]
	}
	return "latest"
}
