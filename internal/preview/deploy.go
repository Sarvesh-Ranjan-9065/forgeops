package preview

import (
	"context"
	"log/slog"
	"strings"
)

// DeployArgs constructs the argument list for helm upgrade --install. Flag
// parsing is terminated with -- before user-controlled arguments.
func DeployArgs(namespace, svc, chartPath, image, host string) []string {
	return []string{
		"upgrade", "--install",
		"--namespace", namespace,
		"--create-namespace",
		"--set-string", "image.repository=" + imageRepo(image),
		"--set-string", "image.tag=" + imageTag(image),
		"--set-string", "ingress.host=" + host,
		"--set", "replicaCount=1",
		"--set", "resources.limits.cpu=500m",
		"--set", "resources.limits.memory=512Mi",
		"--wait", "--timeout", "120s",
		"--",
		svc,
		chartPath,
	}
}

// Deploy installs or upgrades the per-PR Helm release into namespace, overriding
// the image and ingress host with PR-specific values. Flag parsing is terminated
// with -- before user-controlled release and path arguments to prevent flag injection.
func Deploy(ctx context.Context, logger *slog.Logger, namespace, svc, chartPath, image, host string) error {
	return run(ctx, logger, "helm", DeployArgs(namespace, svc, chartPath, image, host)...)
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
