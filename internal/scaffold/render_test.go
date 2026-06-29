package scaffold

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// update regenerates the golden files when set: go test ./internal/scaffold
// -update
var update = flag.Bool("update", false, "update golden files in testdata")

func TestRenderGolden(t *testing.T) {
	spec := ServiceSpec{Name: "demo", Lang: "go", Port: 8080, Image: "demo"}
	outDir := t.TempDir()
	r := NewRenderer()
	require.NoError(t, r.Render(spec, outDir))
	goldenRoot := filepath.Join("testdata", "golden")
	renderedRoot := filepath.Join(outDir, spec.Name)
	if *update {
		require.NoError(t, os.RemoveAll(goldenRoot))
		require.NoError(t, copyTree(renderedRoot, goldenRoot))
		return
	}
	err := filepath.WalkDir(renderedRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(renderedRoot, path)
		if err != nil {
			return err
		}
		got, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		want, err := os.ReadFile(filepath.Join(goldenRoot, rel))
		require.NoErrorf(t, err, "missing golden file for %s (run go test -update)", rel)
		require.Equalf(t, string(want), string(got), "rendered %s differs from golden", rel)
		return nil
	})
	require.NoError(t, err)
}
func TestServiceSpecValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := ServiceSpec{Name: "demo", Lang: "go", Port: 8080, Image: "demo"}
		require.NoError(t, s.Validate())
	})
	t.Run("bad name", func(t *testing.T) {
		s := ServiceSpec{Name: "Demo_Service", Lang: "go", Port: 8080, Image: "x"}
		require.Error(t, s.Validate())
	})
	t.Run("unsupported lang", func(t *testing.T) {
		s := ServiceSpec{Name: "demo", Lang: "rust", Port: 8080, Image: "x"}
		require.Error(t, s.Validate())
	})
}

// copyTree copies the file tree rooted at src into dst, creating directories as
// needed. It is used only by the test's -update path.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}
