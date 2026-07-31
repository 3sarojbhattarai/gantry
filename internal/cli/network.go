package cli

import (
	"context"
	"fmt"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newNetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage networks",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newNetworkCreateCmd(),
		newNetworkRmCmd(),
		newNetworkConnectCmd(),
		newNetworkDisconnectCmd(),
	)
	return cmd
}

func newNetworkCreateCmd() *cobra.Command {
	var driver string
	var internal bool
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				id, err := cli.CreateNetwork(ctx, docker.CreateNetworkOptions{
					Name:     args[0],
					Driver:   driver,
					Internal: internal,
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), id)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&driver, "driver", "d", "bridge", "Network driver")
	cmd.Flags().BoolVar(&internal, "internal", false, "Restrict external access to the network")
	return cmd
}

func newNetworkRmCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "rm <network>...",
		Short: "Remove one or more networks",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return forEachTarget(cmd, args, func(ctx context.Context, cli docker.Client, id string) error {
				return cli.RemoveNetwork(ctx, id, docker.RemoveNetworkOptions{Consent: consentFrom(force)})
			})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm removal")
	return cmd
}

func newNetworkConnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect <network> <container>",
		Short: "Connect a container to a network",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				return cli.ConnectNetwork(ctx, args[0], args[1])
			})
		},
	}
}

func newNetworkDisconnectCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "disconnect <network> <container>",
		Short: "Disconnect a container from a network",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				return cli.DisconnectNetwork(ctx, args[0], args[1], force)
			})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force the disconnect")
	return cmd
}
