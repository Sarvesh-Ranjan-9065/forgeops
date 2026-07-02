package cli

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/config"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/gh"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/log"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/scaffold"
)

var (
	newName    string
	newLang    string
	newOutput  string
	newPush    bool
	newPrivate bool
)

// newCmd scaffolds a new service (Helm chart, Dockerfile, CI workflow) to disk.
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Scaffold a new service to disk",
	RunE: func(cmd *cobra.Command, _ []string) error {
		level := "info"
		if verbose {
			level = "debug"
		}
		logger := log.New(level)
		spec := scaffold.ServiceSpec{Name: newName, Lang: newLang}
		spec.Default()
		if err := spec.Validate(); err != nil {
			return err
		}
		renderer := scaffold.NewRenderer()
		if newPush {
			return pushToGitHub(cmd.Context(), logger, renderer, spec, newPrivate)
		}
		if err := renderer.Render(spec, newOutput); err != nil {
			return err
		}
		logger.Info("scaffolded service", "name", spec.Name, "output", newOutput)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&newName, "name", "", "service name (required, lowercase DNS label)")
	newCmd.Flags().StringVar(&newLang, "lang", "go", "implementation language")
	newCmd.Flags().StringVar(&newOutput, "output", "./out", "output directory")
	newCmd.Flags().BoolVar(&newPush, "push", false, "create the GitHub repo and push the scaffold")
	newCmd.Flags().BoolVar(&newPrivate, "private", false, "create the GitHub repo as private")
	_ = newCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(newCmd)
}

// pushToGitHub renders the scaffold in memory and pushes it to a new (or
// existing) GitHub repository named after the service.
func pushToGitHub(ctx context.Context, logger *slog.Logger, r *scaffold.Renderer, spec scaffold.ServiceSpec, private bool) error {
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	files, err := r.RenderMap(spec)
	if err != nil {
		return err
	}
	client := gh.NewClient(ctx, cfg.GitHubToken)
	if err := gh.CreateRepo(ctx, client, spec.Name, private); err != nil {
		return err
	}
	if err := gh.PushTree(ctx, client, cfg.GitHubOwner, spec.Name, "chore: scaffold service via forge", files); err != nil {
		return err
	}
	logger.Info("pushed service to github",
		"owner", cfg.GitHubOwner, "repo", spec.Name, "token", cfg.RedactedToken())
	return nil
}
