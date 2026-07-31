package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// mobyExec adapts a hijacked exec attachment to the ExecSession interface.
type mobyExec struct {
	conn   types.HijackedResponse
	execID string
	cli    execResizer
	ctx    context.Context
}

// execResizer is the slice of the SDK client mobyExec needs, so tests don't
// need a full client.
type execResizer interface {
	ContainerExecResize(ctx context.Context, execID string, options container.ResizeOptions) error
}

func (e *mobyExec) Read(p []byte) (int, error)  { return e.conn.Reader.Read(p) }
func (e *mobyExec) Write(p []byte) (int, error) { return e.conn.Conn.Write(p) }

func (e *mobyExec) Close() error {
	e.conn.Close()
	return nil
}

func (e *mobyExec) Resize(rows, cols uint) error {
	return e.cli.ContainerExecResize(e.ctx, e.execID, container.ResizeOptions{Height: rows, Width: cols})
}

func (m *mobyClient) ContainerExec(ctx context.Context, id string, opts ExecOptions) (ExecSession, error) {
	created, err := m.cli.ContainerExecCreate(ctx, id, container.ExecOptions{
		Cmd:          opts.Cmd,
		Tty:          opts.TTY,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Env:          opts.Env,
		WorkingDir:   opts.WorkingDir,
		User:         opts.User,
	})
	if err != nil {
		return nil, fmt.Errorf("gantry: creating exec in %s: %w", id, mapErr(err))
	}
	att, err := m.cli.ContainerExecAttach(ctx, created.ID, container.ExecAttachOptions{Tty: opts.TTY})
	if err != nil {
		return nil, fmt.Errorf("gantry: attaching exec in %s: %w", id, mapErr(err))
	}
	return &mobyExec{conn: att, execID: created.ID, cli: m.cli, ctx: ctx}, nil
}
