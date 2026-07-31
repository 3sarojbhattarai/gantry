package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
)

// runErr executes the root command and returns output plus any error, without
// failing the test (so error paths can be asserted).
func runErr(args ...string) (string, error) {
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return buf.String(), err
}

func TestStartStopCommands(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "c1", Names: []string{"c1"}, State: "exited"}}}
	useFake(t, f)

	if _, err := runErr("start", "c1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if f.Containers[0].State != "running" {
		t.Fatalf("state = %q, want running", f.Containers[0].State)
	}
	if _, err := runErr("stop", "c1"); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if f.Containers[0].State != "exited" {
		t.Fatalf("state = %q, want exited", f.Containers[0].State)
	}
}

func TestRmRequiresForce(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "c1", Names: []string{"c1"}, State: "exited"}}}
	useFake(t, f)

	// Without --force the engine refuses and the command reports it.
	out, err := runErr("rm", "c1")
	if err == nil {
		t.Fatal("rm without --force should fail")
	}
	if !strings.Contains(out, "destructive") {
		t.Fatalf("expected a consent hint, got:\n%s", out)
	}
	if len(f.Containers) != 1 {
		t.Fatal("container removed without --force")
	}

	// With --force it goes through.
	if _, err := runErr("rm", "-f", "c1"); err != nil {
		t.Fatalf("rm -f: %v", err)
	}
	if len(f.Containers) != 0 {
		t.Fatal("container not removed with --force")
	}
}

func TestPruneDryRun(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{
		{ID: "keep", State: "running"},
		{ID: "drop", State: "exited"},
	}}
	useFake(t, f)

	out, err := runErr("prune", "containers", "--dry-run")
	if err != nil {
		t.Fatalf("prune dry-run: %v", err)
	}
	if !strings.Contains(out, "Would remove 1 containers") || !strings.Contains(out, "drop") {
		t.Fatalf("unexpected dry-run output:\n%s", out)
	}
	if len(f.Containers) != 2 {
		t.Fatal("dry-run removed containers")
	}
}

func TestPruneRefusesWithoutForce(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "drop", State: "exited"}}}
	useFake(t, f)

	out, err := runErr("prune", "containers")
	if err == nil {
		t.Fatal("prune without --force should fail")
	}
	if !strings.Contains(out, "refusing to prune") {
		t.Fatalf("expected refusal message, got:\n%s", out)
	}
	if len(f.Containers) != 1 {
		t.Fatal("prune ran without --force")
	}
}

func TestNetworkCreate(t *testing.T) {
	f := &fakedocker.Fake{}
	useFake(t, f)
	out, err := runErr("network", "create", "mynet")
	if err != nil {
		t.Fatalf("network create: %v", err)
	}
	if !strings.Contains(out, "net-mynet") {
		t.Fatalf("expected created id in output:\n%s", out)
	}
}
