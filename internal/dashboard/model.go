// Package dashboard serves a read-only web view of ForgeOps preview
// environments. It exposes a small JSON API and an embedded single-page UI,
// performs no mutations, and requires no authentication.
package dashboard

import "time"

// EnvView is the read-only representation of one preview environment as rendered
// by the dashboard.
type EnvView struct {
	PR               int       `json:"pr"`
	Service          string    `json:"service"`
	Namespace        string    `json:"namespace"`
	URL              string    `json:"url"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	AgeSeconds       int64     `json:"ageSeconds"`
	ExpiresInSeconds int64     `json:"expiresInSeconds"`
}

// Summary is the aggregate header shown above the environment list.
type Summary struct {
	Active      int       `json:"active"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// APIResponse is the payload returned by the dashboard data endpoint.
type APIResponse struct {
	Summary      Summary   `json:"summary"`
	Environments []EnvView `json:"environments"`
}