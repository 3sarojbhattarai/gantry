package cli

import (
	"context"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <container>...",
		Short: "Start one or more stopped containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return forEachTarget(cmd, args, func(ctx context.Context, cli docker.Client, id string) error {
				return cli.StartContainer(ctx, id)
			})
		},
	}
}

func newStopCmd() *cobra.Command {
	var timeout int
	cmd := &cobra.Command{
		Use:   "stop <container>...",
		Short: "Stop one or more running containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := stopOptions(timeout)
			return forEachTarget(cmd, args, func(ctx context.Context, cli docker.Client, id string) error {
				return cli.StopContainer(ctx, id, opts)
			})
		},
	}
	cmd.Flags().IntVarP(&timeout, "time", "t", -1, "Seconds to wait before killing (-1 = daemon default)")
	return cmd
}

func newRestartCmd() *cobra.Command {
	var timeout int
	cmd := &cobra.Command{
		Use:   "restart <container>...",
		Short: "Restart one or more containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := stopOptions(timeout)
			return forEachTarget(cmd, args, func(ctx context.Context, cli docker.Client, id string) error {
				return cli.RestartContainer(ctx, id, opts)
			})
		},
	}
	cmd.Flags().IntVarP(&timeout, "time", "t", -1, "Seconds to wait before killing (-1 = daemon default)")
	return cmd
}

func newKillCmd() *cobra.Command {
	var signal string
	cmd := &cobra.Command{
		Use:   "kill <container>...",
		Short: "Kill one or more running containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return forEachTarget(cmd, args, func(ctx context.Context, cli docker.Client, id string) error {
				return cli.KillContainer(ctx, id, signal)
			})
		},
	}
	cmd.Flags().StringVarP(&signal, "signal", "s", "KILL", "Signal to send")
	return cmd
}

// stopOptions turns a --time flag (-1 sentinel = default) into StopOptions.
func stopOptions(timeout int) docker.StopOptions {
	if timeout < 0 {
		return docker.StopOptions{}
	}
	t := timeout
	return docker.StopOptions{Timeout: &t}
}
