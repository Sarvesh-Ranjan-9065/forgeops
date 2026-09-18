package preview

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeployArgs(t *testing.T) {
	t.Run("safe argument list with flag terminator", func(t *testing.T) {
		args := DeployArgs("pr-1-payments", "payments", "/tmp/helm", "localhost:5000/payments:pr-1", "pr-1-payments.127-0-0-1.nip.io")

		// Verify -- appears before the positional arguments
		dashDashIdx := -1
		for i, arg := range args {
			if arg == "--" {
				dashDashIdx = i
				break
			}
		}
		require.Greater(t, dashDashIdx, 0, "expected -- flag terminator in args")
		require.Equal(t, "payments", args[dashDashIdx+1])
		require.Equal(t, "/tmp/helm", args[dashDashIdx+2])

		// Verify --set-string is used for user-controlled strings
		hasSetStringRepo := false
		hasSetStringHost := false
		for i, arg := range args {
			if arg == "--set-string" && i+1 < len(args) {
				if args[i+1] == "image.repository=localhost:5000/payments" {
					hasSetStringRepo = true
				}
				if args[i+1] == "ingress.host=pr-1-payments.127-0-0-1.nip.io" {
					hasSetStringHost = true
				}
			}
		}
		require.True(t, hasSetStringRepo, "expected --set-string for image.repository")
		require.True(t, hasSetStringHost, "expected --set-string for ingress.host")
	})

	t.Run("service name starting with hyphen is safely placed after flag terminator", func(t *testing.T) {
		maliciousSvc := "--post-renderer"
		args := DeployArgs("pr-1-test", maliciousSvc, "/tmp/helm", "localhost:5000/test:pr-1", "host.nip.io")

		dashDashIdx := -1
		for i, arg := range args {
			if arg == "--" {
				dashDashIdx = i
				break
			}
		}
		require.Greater(t, dashDashIdx, 0)
		require.Equal(t, maliciousSvc, args[dashDashIdx+1])
	})
}
