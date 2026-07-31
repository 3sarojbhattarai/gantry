package fakedocker

import (
	"context"
	"io"
	"strings"

	"github.com/3sarojbhattarai/gantry/internal/docker"
)

// fakeExec is a no-op exec session: writes are captured, reads report EOF, and
// resizes are recorded. Enough to test the plumbing without a real process.
type fakeExec struct {
	Written strings.Builder
	Rows    uint
	Cols    uint
}

func (e *fakeExec) Read(_ []byte) (int, error)  { return 0, io.EOF }
func (e *fakeExec) Write(p []byte) (int, error) { return e.Written.Write(p) }
func (e *fakeExec) Resize(rows, cols uint) error {
	e.Rows, e.Cols = rows, cols
	return nil
}
func (e *fakeExec) Close() error { return nil }

func (f *Fake) ContainerExec(_ context.Context, id string, _ docker.ExecOptions) (docker.ExecSession, error) {
	if f.ExecErr != nil {
		return nil, f.ExecErr
	}
	for _, c := range f.Containers {
		if c.ID == id || (len(c.Names) > 0 && c.Names[0] == id) {
			return &fakeExec{}, nil
		}
	}
	return nil, docker.ErrNotFound
}

func (f *Fake) CreateContainer(_ context.Context, spec docker.CreateSpec, start bool) (string, error) {
	if f.CreateErr != nil {
		return "", f.CreateErr
	}
	state := "created"
	if start {
		state = "running"
	}
	id := "created-" + spec.Name
	f.Containers = append(f.Containers, docker.Container{
		ID:      id,
		Names:   []string{spec.Name},
		Image:   spec.Image,
		Command: strings.Join(spec.Command, " "),
		State:   state,
		Status:  "Created",
	})
	return id, nil
}

func (f *Fake) SpecFromContainer(_ context.Context, id string) (docker.CreateSpec, error) {
	for _, c := range f.Containers {
		if c.ID == id || (len(c.Names) > 0 && c.Names[0] == id) {
			name := ""
			if len(c.Names) > 0 {
				name = c.Names[0]
			}
			cmd := []string{}
			if c.Command != "" {
				cmd = strings.Fields(c.Command)
			}
			return docker.CreateSpec{Image: c.Image, Name: name, Command: cmd}, nil
		}
	}
	return docker.CreateSpec{}, docker.ErrNotFound
}
