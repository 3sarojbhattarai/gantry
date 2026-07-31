// Package cli holds gantry's one-shot commands, built on the docker engine.
// Phase 0 ships only the root command and version; ps/images/logs land in
// Phase 1.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is the build version, overridden at release time via
// -ldflags "-X github.com/3sarojbhattarai/gantry/internal/cli.version=vX.Y.Z".
var version = "dev"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gantry",
		Short: "A terminal and web UI for managing Docker",
		Long: "gantry is a lazydocker-style tool for inspecting and managing a " +
			"local Docker daemon from the terminal or a browser.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("gantry {{.Version}}\n")
	root.AddCommand(newVersionCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the gantry version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "gantry %s\n", version)
			return nil
		},
	}
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gantry:", err)
		os.Exit(1)
	}
}
