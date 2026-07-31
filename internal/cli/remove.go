package cli

import (
	"context"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	var force, volumes bool
	cmd := &cobra.Command{
		Use:   "rm <container>...",
		Short: "Remove one or more containers",
		Long:  "Remove containers. This is destructive: pass --force to confirm (and to remove a running container).",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return forEachTarget(cmd, args, func(ctx context.Context, cli docker.Client, id string) error {
				return cli.RemoveContainer(ctx, id, docker.RemoveContainerOptions{
					Consent:       consentFrom(force),
					Force:         force,
					RemoveVolumes: volumes,
				})
			})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm removal (and force-remove running containers)")
	cmd.Flags().BoolVarP(&volumes, "volumes", "v", false, "Also remove anonymous volumes")
	return cmd
}

func newRmiCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "rmi <image>...",
		Short: "Remove one or more images",
		Long:  "Remove images. This is destructive: pass --force to confirm.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return forEachTarget(cmd, args, func(ctx context.Context, cli docker.Client, id string) error {
				return cli.RemoveImage(ctx, id, docker.RemoveImageOptions{
					Consent:       consentFrom(force),
					Force:         force,
					PruneChildren: true,
				})
			})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm removal")
	return cmd
}
