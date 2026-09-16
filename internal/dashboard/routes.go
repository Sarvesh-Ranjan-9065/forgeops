package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed static/index.html
var indexHTML []byte

// Routes returns an http.Handler serving the dashboard UI at / and its JSON API
// at /api/environments. Mount it under a prefix such as /dashboard.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/api/environments", h.data)
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})
	r.Get("/*", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})
	return r
}