package cli

import (
	"context"
	"os"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/termexec"
	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	var tty bool
	cmd := &cobra.Command{
		Use:   "exec <container> <command> [args...]",
		Short: "Run a command in a running container",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				sess, err := cli.ContainerExec(ctx, args[0], docker.ExecOptions{Cmd: args[1:], TTY: tty})
				if err != nil {
					return err
				}
				defer sess.Close()
				return termexec.Attach(sess, os.Stdin, os.Stdout)
			})
		},
	}
	cmd.Flags().BoolVarP(&tty, "tty", "t", true, "Allocate a pseudo-TTY")
	return cmd
}
