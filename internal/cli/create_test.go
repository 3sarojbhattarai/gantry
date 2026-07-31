package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
)

func TestCreateFromExportRun(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{
		{ID: "a", Names: []string{"web"}, Image: "nginx", Command: "nginx -g daemon"},
	}}
	useFake(t, f)

	out, err := runErr("create", "--from", "web", "--export", "run")
	if err != nil {
		t.Fatalf("create --export run: %v", err)
	}
	if !strings.Contains(out, "docker run") || !strings.Contains(out, "nginx") {
		t.Fatalf("unexpected export:\n%s", out)
	}
}

func TestCreateFromActuallyCreates(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{
		{ID: "a", Names: []string{"web"}, Image: "nginx"},
	}}
	useFake(t, f)

	out, err := runErr("create", "--from", "web")
	if err != nil {
		t.Fatalf("create --from: %v", err)
	}
	if !strings.Contains(out, "created-web") {
		t.Fatalf("expected new id in output:\n%s", out)
	}
	if len(f.Containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(f.Containers))
	}
}

func TestCreateFromFile(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(spec, []byte("image: redis:7\nname: cache\nports:\n  - host: \"6379\"\n    container: 6379\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f := &fakedocker.Fake{}
	useFake(t, f)

	if _, err := runErr("create", "--file", spec); err != nil {
		t.Fatalf("create --file: %v", err)
	}
	if len(f.Containers) != 1 || f.Containers[0].Image != "redis:7" {
		t.Fatalf("container not created from file: %+v", f.Containers)
	}
}

func TestCreateRequiresOneSource(t *testing.T) {
	useFake(t, &fakedocker.Fake{})
	if _, err := runErr("create"); err == nil {
		t.Fatal("create with neither --file nor --from should fail")
	}
}
