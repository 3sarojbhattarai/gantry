package tui

import (
	"context"
	"io"
	"os"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/termexec"
)

// execFinishedMsg reports that an interactive exec (which suspended the TUI)
// has returned.
type execFinishedMsg struct{ err error }

// execCommand is a tea.ExecCommand: Bubbletea releases the terminal, runs it,
// and restores the terminal afterward. It opens a shell in the container and
// hands the real terminal over via termexec.
type execCommand struct {
	ctx    context.Context
	client docker.Client
	id     string
	stdin  io.Reader
	stdout io.Writer
}

func (c *execCommand) SetStdin(r io.Reader)  { c.stdin = r }
func (c *execCommand) SetStdout(w io.Writer) { c.stdout = w }
func (c *execCommand) SetStderr(_ io.Writer) {}

func (c *execCommand) Run() error {
	sess, err := c.client.ContainerExec(c.ctx, c.id, docker.ExecOptions{
		Cmd: []string{"/bin/sh"},
		TTY: true,
	})
	if err != nil {
		return err
	}
	defer sess.Close()

	in, ok := c.stdin.(*os.File)
	if !ok {
		in = os.Stdin
	}
	out, ok := c.stdout.(*os.File)
	if !ok {
		out = os.Stdout
	}
	return termexec.Attach(sess, in, out)
}
