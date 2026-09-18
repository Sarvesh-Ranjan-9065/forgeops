package preview

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeardownArgs(t *testing.T) {
	t.Run("terminates flags before release name", func(t *testing.T) {
		args := TeardownArgs("pr-42-orders", "orders")

		dashDashIdx := -1
		for i, arg := range args {
			if arg == "--" {
				dashDashIdx = i
				break
			}
		}
		require.Greater(t, dashDashIdx, 0, "expected -- flag terminator in args")
		require.Equal(t, "orders", args[dashDashIdx+1])
	})

	t.Run("handles release name starting with hyphen safely", func(t *testing.T) {
		maliciousSvc := "--dry-run"
		args := TeardownArgs("pr-42-orders", maliciousSvc)

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
