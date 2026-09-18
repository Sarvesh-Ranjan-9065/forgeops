package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCmd(t *testing.T) {
	t.Run("missing name flag returns error", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"new"})
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("invalid name format returns error", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"new", "--name", "INVALID_NAME!"})
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("valid name scaffolds files to directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		cmd := rootCmd
		cmd.SetArgs([]string{"new", "--name", "testservice", "--output", tmpDir})
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify files exist in output directory
		serviceDir := filepath.Join(tmpDir, "testservice")
		require.DirExists(t, serviceDir)
		require.FileExists(t, filepath.Join(serviceDir, "Dockerfile"))
		require.FileExists(t, filepath.Join(serviceDir, "helm", "Chart.yaml"))
	})
}
