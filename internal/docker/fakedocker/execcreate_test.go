package fakedocker_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
)

func TestFakeCreateContainer(t *testing.T) {
	f := &fakedocker.Fake{}
	id, err := f.CreateContainer(context.Background(), docker.CreateSpec{Name: "web", Image: "nginx"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if id != "created-web" || len(f.Containers) != 1 || f.Containers[0].State != "running" {
		t.Fatalf("create: id=%q containers=%+v", id, f.Containers)
	}
}

func TestFakeSpecFromContainer(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{
		{ID: "a", Names: []string{"web"}, Image: "nginx", Command: "nginx -g daemon"},
	}}
	spec, err := f.SpecFromContainer(context.Background(), "a")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Image != "nginx" || spec.Name != "web" || len(spec.Command) != 3 {
		t.Fatalf("spec = %+v", spec)
	}
	if _, err := f.SpecFromContainer(context.Background(), "missing"); !errors.Is(err, docker.ErrNotFound) {
		t.Fatalf("missing: got %v", err)
	}
}

func TestFakeExec(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", Names: []string{"web"}}}}
	sess, err := f.ContainerExec(context.Background(), "a", docker.ExecOptions{Cmd: []string{"sh"}, TTY: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sess.Write([]byte("echo hi\n")); err != nil {
		t.Fatal(err)
	}
	if err := sess.Resize(24, 80); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4)
	if _, err := sess.Read(buf); !errors.Is(err, io.EOF) {
		t.Fatalf("read = %v, want EOF", err)
	}
	if err := sess.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := f.ContainerExec(context.Background(), "missing", docker.ExecOptions{}); !errors.Is(err, docker.ErrNotFound) {
		t.Fatalf("exec missing: got %v", err)
	}
}
