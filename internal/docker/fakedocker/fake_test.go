package fakedocker_test

import (
	"context"
	"errors"
	"io"
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

func TestFakeContainerLogs(t *testing.T) {
	f := &fakedocker.Fake{Logs: map[string]string{"c1": "hello\n"}}
	rc, err := f.ContainerLogs(context.Background(), "c1", docker.LogOptions{})
	if err != nil {
		t.Fatalf("ContainerLogs: %v", err)
	}
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	if string(b) != "hello\n" {
		t.Fatalf("logs = %q, want %q", b, "hello\n")
	}
}

func TestFakeContainerStatsStream(t *testing.T) {
	f := &fakedocker.Fake{StatsSamples: []docker.Stats{{CPUPercent: 1}, {CPUPercent: 2}}}
	ch, err := f.ContainerStats(context.Background(), "c1")
	if err != nil {
		t.Fatalf("ContainerStats: %v", err)
	}
	var got int
	for range ch {
		got++
	}
	if got != 2 {
		t.Fatalf("received %d samples, want 2", got)
	}
}

func TestFakeEventsStreamStopsOnCancel(t *testing.T) {
	f := &fakedocker.Fake{EventStream: []docker.Event{{Action: "start"}, {Action: "die"}}}
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := f.Events(ctx)
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if ev, ok := <-ch; !ok || ev.Action != "start" {
		t.Fatalf("first event = %+v, ok=%v", ev, ok)
	}
	cancel()
	// Draining after cancel must terminate (channel closes), not block forever.
	for range ch {
	}
}
