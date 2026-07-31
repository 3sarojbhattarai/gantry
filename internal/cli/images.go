package cli

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "List images",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				imgs, err := cli.ListImages(ctx)
				if err != nil {
					return err
				}
				return renderImages(cmd.OutOrStdout(), imgs)
			})
		},
	}
	return cmd
}

func renderImages(w io.Writer, imgs []docker.ImageSummary) error {
	tw := tabwriter.NewWriter(w, 0, 2, 3, ' ', 0)
	fmt.Fprintln(tw, "REPOSITORY\tTAG\tIMAGE ID\tSIZE")
	for _, im := range imgs {
		id, size := shortID(im.ID), humanSize(im.Size)
		if len(im.RepoTags) == 0 {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "<none>", "<none>", id, size)
			continue
		}
		// Docker prints one row per repository:tag reference.
		for _, ref := range im.RepoTags {
			repo, tag := splitRepoTag(ref)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", repo, tag, id, size)
		}
	}
	return tw.Flush()
}
