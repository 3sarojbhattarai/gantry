package cli

import (
	"context"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage volumes",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newVolumeRmCmd())
	return cmd
}

func newVolumeRmCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "rm <volume>...",
		Short: "Remove one or more volumes",
		Long:  "Remove volumes. This is destructive: pass --force to confirm.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return forEachTarget(cmd, args, func(ctx context.Context, cli docker.Client, name string) error {
				return cli.RemoveVolume(ctx, name, docker.RemoveVolumeOptions{Consent: consentFrom(force), Force: force})
			})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm removal")
	return cmd
}
