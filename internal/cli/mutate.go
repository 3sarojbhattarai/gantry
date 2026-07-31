package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
)

// errSilent marks an error whose details a command has already printed itself,
// so Execute should set a non-zero exit code without printing anything more.
var errSilent = errors.New("")

// consentFrom turns a --force flag into engine consent. Without --force the
// zero (ungranted) Consent is returned, and destructive engine calls fail with
// ErrConsentRequired.
func consentFrom(force bool) docker.Consent {
	if force {
		return docker.Confirm()
	}
	return docker.Consent{}
}

// forEachTarget runs fn against each argument, printing the id on success and a
// per-target message on failure. It returns errSilent if any target failed.
func forEachTarget(cmd *cobra.Command, args []string, fn func(context.Context, docker.Client, string) error) error {
	return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
		failed := false
		for _, id := range args {
			err := fn(ctx, cli, id)
			switch {
			case err == nil:
				fmt.Fprintln(cmd.OutOrStdout(), id)
			case errors.Is(err, docker.ErrConsentRequired):
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: refused — destructive, pass --force to confirm\n", id)
				failed = true
			default:
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: %v\n", id, err)
				failed = true
			}
		}
		if failed {
			return errSilent
		}
		return nil
	})
}
