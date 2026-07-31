package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newPSCmd() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List containers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				cs, err := cli.ListContainers(ctx, all)
				if err != nil {
					return err
				}
				return renderContainers(cmd.OutOrStdout(), cs)
			})
		},
	}
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all containers (default shows just running)")
	return cmd
}

func renderContainers(w io.Writer, cs []docker.Container) error {
	tw := tabwriter.NewWriter(w, 0, 2, 3, ' ', 0)
	fmt.Fprintln(tw, "CONTAINER ID\tIMAGE\tCOMMAND\tSTATUS\tPORTS\tNAMES")
	for _, c := range cs {
		fmt.Fprintf(tw, "%s\t%s\t%q\t%s\t%s\t%s\n",
			shortID(c.ID),
			c.Image,
			truncate(c.Command, 20),
			c.Status,
			formatPorts(c.Ports),
			strings.Join(c.Names, ", "),
		)
	}
	return tw.Flush()
}
