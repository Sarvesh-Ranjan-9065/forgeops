// Package cli defines the ForgeOps command-line interface and wires together the
// individual subcommands. In Phase 0 it provides only the root command and
// shared persistent flags; subcommands are added in later phases.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// verbose enables debug-level output across all forge subcommands. It is bound
// to the persistent --verbose flag on the root command.
var verbose bool

// rootCmd is the base command for the forge CLI. All other commands are attached
// to it as subcommands.
var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "ForgeOps scaffolds Go services and manages PR preview environments",
	Long: "forge is the ForgeOps control plane CLI. It scaffolds new Go services and, " +
		"in later phases, reconciles GitHub pull requests into ephemeral preview environments.",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

// Execute runs the root command. It writes any error to standard error and exits
// the process with a non-zero status, so callers in main do not need to handle
// errors themselves.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
