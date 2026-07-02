package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// templateRoot is the directory within the embedded filesystem that holds the
// service templates.
const templateRoot = "templates"

// Renderer renders the embedded service templates against a ServiceSpec and
// writes the resulting files to disk.
type Renderer struct {
	fsys fs.FS
}

// NewRenderer returns a Renderer backed by the embedded template filesystem.
func NewRenderer() *Renderer {
	return &Renderer{fsys: TemplatesFS}
}

// Render executes every embedded template against spec and writes the output
// tree under dest/<spec.Name>. Templates use the "[[" and "]]" delimiters so
// that Helm and GitHub Actions " " expressions pass through untouched. The
// ".tmpl" suffix is stripped from each output filename.
func (r *Renderer) Render(spec ServiceSpec, dest string) error {
	if err := spec.Validate(); err != nil {
		return fmt.Errorf("invalid service spec: %w", err)
	}
	root := filepath.Join(dest, spec.Name)
	return fs.WalkDir(r.fsys, templateRoot, func(path string, d fs.DirEntry,
		walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, templateRoot+"/")
		outRel := strings.TrimSuffix(rel, ".tmpl")
		outPath := filepath.Join(root, filepath.FromSlash(outRel))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("create dir for %s: %w", outPath, err)
		}
		return r.renderFile(path, outPath, spec)
	})
}

// renderFile renders a single template and writes the result to outPath.
func (r *Renderer) renderFile(srcPath, outPath string, spec ServiceSpec) error {
	b, err := r.renderBytes(srcPath, spec)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}

// renderBytes parses and executes a single template file and returns the result.
func (r *Renderer) renderBytes(srcPath string, spec ServiceSpec) ([]byte, error) {
	raw, err := fs.ReadFile(r.fsys, srcPath)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", srcPath, err)
	}
	tmpl, err := template.New(filepath.Base(srcPath)).Delims("[[", "]]").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", srcPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, spec); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", srcPath, err)
	}
	return buf.Bytes(), nil
}

// RenderMap renders all templates against spec and returns a map of
// repository-relative paths to file contents, without writing to disk. It backs
// the GitHub push path.
func (r *Renderer) RenderMap(spec ServiceSpec) (map[string][]byte, error) {
	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid service spec: %w", err)
	}
	out := make(map[string][]byte)
	err := fs.WalkDir(r.fsys, templateRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimSuffix(strings.TrimPrefix(path, templateRoot+"/"), ".tmpl")
		b, err := r.renderBytes(path, spec)
		if err != nil {
			return err
		}
		out[rel] = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
