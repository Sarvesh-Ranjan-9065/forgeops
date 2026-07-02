// Package config loads ForgeOps runtime configuration. Secrets are read only
// from the environment, never from flags or files.
package config

import (
	"errors"
	"os"
)

// Config holds runtime configuration for ForgeOps.
type Config struct {
	// GitHubToken authenticates GitHub API calls. Source: FORGE_GITHUB_TOKEN.
	GitHubToken string
	// GitHubOwner is the account or org that owns pushed repositories.
	// Source: FORGE_GITHUB_OWNER.
	GitHubOwner string
}

// FromEnv builds a Config from environment variables. Both FORGE_GITHUB_TOKEN
// and FORGE_GITHUB_OWNER are required for GitHub operations.
func FromEnv() (Config, error) {
	c := Config{
		GitHubToken: os.Getenv("FORGE_GITHUB_TOKEN"),
		GitHubOwner: os.Getenv("FORGE_GITHUB_OWNER"),
	}
	if c.GitHubToken == "" {
		return Config{}, errors.New("FORGE_GITHUB_TOKEN is required")
	}
	if c.GitHubOwner == "" {
		return Config{}, errors.New("FORGE_GITHUB_OWNER is required")
	}
	return c, nil
}

// RedactedToken returns a safe-to-log representation of the token that reveals
// at most the last four characters.
func (c Config) RedactedToken() string {
	if len(c.GitHubToken) <= 4 {
		return "****"
	}
	return "****" + c.GitHubToken[len(c.GitHubToken)-4:]
}
