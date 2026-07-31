// Package cli holds gantry's one-shot commands, built on the docker engine.
// Running gantry with no subcommand launches the terminal UI.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/tui"
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
			"local Docker daemon from the terminal or a browser.\n\n" +
			"Run with no subcommand to launch the terminal UI.",
		Version:       version,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				return tui.Run(ctx, cli)
			})
		},
	}
	root.SetVersionTemplate("gantry {{.Version}}\n")
	root.AddCommand(
		newVersionCmd(),
		newPSCmd(),
		newImagesCmd(),
		newLogsCmd(),
		newStartCmd(),
		newStopCmd(),
		newRestartCmd(),
		newKillCmd(),
		newRmCmd(),
		newRmiCmd(),
		newVolumeCmd(),
		newNetworkCmd(),
		newPruneCmd(),
		newExecCmd(),
		newCreateCmd(),
		newServeCmd(),
	)
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

// Execute runs the root command and exits non-zero on error. The context is
// cancelled on SIGINT/SIGTERM so streaming commands (e.g. `logs -f`) shut down
// cleanly when interrupted.
func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		// errSilent means a command already printed its own diagnostics; just
		// exit non-zero without a duplicate message.
		if !errors.Is(err, errSilent) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
