package fakedocker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
)

// This test establishes the pattern Phase 1's engine/CLI tests follow: drive a
// consumer against fakedocker.Fake, no daemon required.
func TestFakeListContainers(t *testing.T) {
	f := &fakedocker.Fake{
		Containers: []docker.Container{
			{ID: "a", State: "running"},
			{ID: "b", State: "exited"},
		},
	}

	all, err := f.ListContainers(context.Background(), true)
	if err != nil {
		t.Fatalf("ListContainers(all=true): %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("all: got %d containers, want 2", len(all))
	}

	running, err := f.ListContainers(context.Background(), false)
	if err != nil {
		t.Fatalf("ListContainers(all=false): %v", err)
	}
	if len(running) != 1 || running[0].ID != "a" {
		t.Fatalf("running: got %+v, want just container a", running)
	}
}

func TestFakeInspectContainerNotFound(t *testing.T) {
	f := &fakedocker.Fake{}
	_, err := f.InspectContainer(context.Background(), "missing")
	if !errors.Is(err, docker.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestFakeErrorInjection(t *testing.T) {
	sentinel := errors.New("boom")
	f := &fakedocker.Fake{ListImagesErr: sentinel}
	if _, err := f.ListImages(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want injected error", err)
	}
}
