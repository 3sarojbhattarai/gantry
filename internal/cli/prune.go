package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

func newPruneCmd() *cobra.Command {
	var force, dryRun bool
	cmd := &cobra.Command{
		Use:       "prune <containers|images|volumes|networks>",
		Short:     "Remove unused resources",
		Long:      "Remove unused resources. Destructive: pass --force to confirm, or --dry-run to preview (containers and images only).",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"containers", "images", "volumes", "networks"},
		RunE: func(cmd *cobra.Command, args []string) error {
			kind := args[0]
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				opts := docker.PruneOptions{Consent: consentFrom(force), DryRun: dryRun}
				var (
					report docker.PruneReport
					err    error
				)
				switch kind {
				case "containers":
					report, err = cli.PruneContainers(ctx, opts)
				case "images":
					report, err = cli.PruneImages(ctx, opts)
				case "volumes":
					report, err = cli.PruneVolumes(ctx, opts)
				case "networks":
					report, err = cli.PruneNetworks(ctx, opts)
				default:
					return fmt.Errorf("gantry: unknown prune target %q (want containers|images|volumes|networks)", kind)
				}

				switch {
				case errors.Is(err, docker.ErrConsentRequired):
					fmt.Fprintln(cmd.ErrOrStderr(), "refusing to prune (destructive); pass --force to confirm or --dry-run to preview")
					return errSilent
				case errors.Is(err, docker.ErrDryRunUnsupported):
					fmt.Fprintf(cmd.ErrOrStderr(), "dry-run is not supported for %s (needs the daemon to compute usage)\n", kind)
					return errSilent
				case err != nil:
					return err
				}
				renderPrune(cmd.OutOrStdout(), kind, report)
				return nil
			})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm the prune")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Report what would be removed without removing it")
	return cmd
}

func renderPrune(w io.Writer, kind string, r docker.PruneReport) {
	verb := "Removed"
	if r.DryRun {
		verb = "Would remove"
	}
	fmt.Fprintf(w, "%s %d %s\n", verb, len(r.Items), kind)
	for _, it := range r.Items {
		fmt.Fprintf(w, "  %s\n", it)
	}
	if r.SpaceReclaimed > 0 {
		label := "reclaimed"
		if r.DryRun {
			label = "reclaimable"
		}
		fmt.Fprintf(w, "Total %s: %s\n", label, humanSize(int64(r.SpaceReclaimed)))
	}
}
