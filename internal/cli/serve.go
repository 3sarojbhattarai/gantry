package cli

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/3sarojbhattarai/gantry/internal/api"
	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the web UI and HTTP API",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				warnIfExposed(cmd.ErrOrStderr(), addr)
				fmt.Fprintf(cmd.OutOrStdout(), "gantry: serving on http://%s\n", addr)
				return api.NewServer(cli).Run(ctx, addr)
			})
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8080",
		"Address to bind (a non-loopback address exposes Docker control to the network)")
	return cmd
}

// warnIfExposed prints a loud warning when addr binds anything other than the
// loopback interface — the server offers full, unauthenticated Docker control.
func warnIfExposed(w io.Writer, addr string) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	switch host {
	case "127.0.0.1", "localhost", "::1":
		return
	}
	fmt.Fprintf(w, "\n  ⚠  WARNING: binding to %q exposes full Docker control to the network.\n", addr)
	fmt.Fprintf(w, "     Anyone who can reach this address can start, stop, and delete your\n")
	fmt.Fprintf(w, "     containers. There is no authentication. Bind 127.0.0.1 unless you are\n")
	fmt.Fprintf(w, "     certain the network is trusted.\n\n")
}
