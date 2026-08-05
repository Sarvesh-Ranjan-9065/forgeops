package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromEnv(t *testing.T) {
	t.Run("missing required variables", func(t *testing.T) {
		_ = os.Unsetenv("FORGE_GITHUB_TOKEN")
		_ = os.Unsetenv("FORGE_GITHUB_OWNER")

		_, err := FromEnv()
		require.Error(t, err)
	})

	t.Run("valid environment variables", func(t *testing.T) {
		t.Setenv("FORGE_GITHUB_TOKEN", "ghp_1234567890abcdef")
		t.Setenv("FORGE_GITHUB_OWNER", "test-org")

		cfg, err := FromEnv()
		require.NoError(t, err)
		require.Equal(t, "ghp_1234567890abcdef", cfg.GitHubToken)
		require.Equal(t, "test-org", cfg.GitHubOwner)
	})
}

func TestRedactedToken(t *testing.T) {
	t.Run("short token", func(t *testing.T) {
		cfg := Config{GitHubToken: "abc"}
		require.Equal(t, "****", cfg.RedactedToken())
	})

	t.Run("long token", func(t *testing.T) {
		cfg := Config{GitHubToken: "ghp_secrettoken1234"}
		require.Equal(t, "****1234", cfg.RedactedToken())
	})
}

func TestWebhookSecret(t *testing.T) {
	t.Run("missing secret", func(t *testing.T) {
		_ = os.Unsetenv("FORGE_WEBHOOK_SECRET")
		_, err := WebhookSecret()
		require.Error(t, err)
	})

	t.Run("valid secret", func(t *testing.T) {
		t.Setenv("FORGE_WEBHOOK_SECRET", "supersecret")
		secret, err := WebhookSecret()
		require.NoError(t, err)
		require.Equal(t, "supersecret", secret)
	})
}
