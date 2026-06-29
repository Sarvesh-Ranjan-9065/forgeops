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

// renderFile parses and executes a single template file, writing the result to
// outPath.
func (r *Renderer) renderFile(srcPath, outPath string, spec ServiceSpec) error {
	raw, err := fs.ReadFile(r.fsys, srcPath)
	if err != nil {
		return fmt.Errorf("read template %s: %w", srcPath, err)
	}
	tmpl, err := template.New(filepath.Base(srcPath)).Delims("[[", "]]").Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", srcPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, spec); err != nil {
		return fmt.Errorf("execute template %s: %w", srcPath, err)
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}
