package preview

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageRef(t *testing.T) {
	ref := ImageRef("localhost:5000", "demo-svc", 42)
	require.Equal(t, "localhost:5000/demo-svc:pr-42", ref)
}

func TestImageRepo(t *testing.T) {
	t.Run("with tag", func(t *testing.T) {
		repo := imageRepo("localhost:5000/demo-svc:pr-42")
		require.Equal(t, "localhost:5000/demo-svc", repo)
	})

	t.Run("without tag", func(t *testing.T) {
		repo := imageRepo("localhost:5000/demo-svc")
		require.Equal(t, "localhost:5000/demo-svc", repo)
	})
}

func TestImageTag(t *testing.T) {
	t.Run("with tag", func(t *testing.T) {
		tag := imageTag("localhost:5000/demo-svc:pr-42")
		require.Equal(t, "pr-42", tag)
	})

	t.Run("without tag defaults to latest", func(t *testing.T) {
		tag := imageTag("localhost:5000/demo-svc")
		require.Equal(t, "latest", tag)
	})
}
