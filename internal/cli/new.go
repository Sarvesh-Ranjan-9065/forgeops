package cli

import (
	"github.com/sarvesh-ranjan-9065/forgeops/internal/log"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/scaffold"
	"github.com/spf13/cobra"
)

var (
	newName   string
	newLang   string
	newOutput string
)

// newCmd scaffolds a new service (Helm chart, Dockerfile, CI workflow) to disk.
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Scaffold a new service to disk",
	RunE: func(_ *cobra.Command, _ []string) error {
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
		if err := scaffold.NewRenderer().Render(spec, newOutput); err != nil {
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
	_ = newCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(newCmd)
}
