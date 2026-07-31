package cli

import (
	"context"
	"io"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var (
		follow     bool
		timestamps bool
		tail       string
	)
	cmd := &cobra.Command{
		Use:   "logs <container>",
		Short: "Fetch the logs of a container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				rc, err := cli.ContainerLogs(ctx, args[0], docker.LogOptions{
					Follow:     follow,
					Timestamps: timestamps,
					Tail:       tail,
				})
				if err != nil {
					return err
				}
				defer rc.Close()

				_, err = io.Copy(cmd.OutOrStdout(), rc)
				// A follow that the user cancels (Ctrl-C) surfaces as a read
				// error on a cancelled context; that's a clean exit, not a
				// failure.
				if err != nil && ctx.Err() != nil {
					return nil
				}
				return err
			})
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().BoolVarP(&timestamps, "timestamps", "t", false, "Show timestamps")
	cmd.Flags().StringVar(&tail, "tail", "all", "Number of lines to show from the end of the logs")
	return cmd
}
