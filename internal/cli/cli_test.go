package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
)

// useFake points the command layer at an in-memory engine for the duration of
// a test.
func useFake(t *testing.T, f *fakedocker.Fake) {
	t.Helper()
	prev := newClient
	newClient = func() (docker.Client, error) { return f, nil }
	t.Cleanup(func() { newClient = prev })
}

// run executes the root command with args and returns combined output.
func run(t *testing.T, args ...string) string {
	t.Helper()
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("execute %v: %v", args, err)
	}
	return buf.String()
}

func TestPSCommand(t *testing.T) {
	useFake(t, &fakedocker.Fake{
		Containers: []docker.Container{
			{
				ID:      "abcdef0123456789",
				Names:   []string{"web"},
				Image:   "nginx:latest",
				Command: "nginx -g daemon off;",
				State:   "running",
				Status:  "Up 2 hours",
				Ports:   []docker.Port{{IP: "0.0.0.0", Public: 8080, Private: 80, Type: "tcp"}},
			},
			{ID: "ffffffff0000", Names: []string{"db"}, Image: "postgres:16", State: "exited", Status: "Exited (0)"},
		},
	})

	// Default: running only.
	out := run(t, "ps")
	if !strings.Contains(out, "abcdef012345") || !strings.Contains(out, "web") {
		t.Errorf("ps missing running container:\n%s", out)
	}
	if !strings.Contains(out, "0.0.0.0:8080->80/tcp") {
		t.Errorf("ps missing port mapping:\n%s", out)
	}
	if strings.Contains(out, "db") {
		t.Errorf("ps (default) should not list the exited container:\n%s", out)
	}

	// -a: all containers.
	if all := run(t, "ps", "-a"); !strings.Contains(all, "db") {
		t.Errorf("ps -a missing exited container:\n%s", all)
	}
}

func TestImagesCommand(t *testing.T) {
	useFake(t, &fakedocker.Fake{
		Images: []docker.ImageSummary{
			{ID: "sha256:1111222233334444", RepoTags: []string{"nginx:latest"}, Size: 187_000_000},
			{ID: "sha256:aaaabbbbcccc", RepoTags: nil, Size: 5_000}, // dangling
		},
	})
	out := run(t, "images")
	if !strings.Contains(out, "nginx") || !strings.Contains(out, "latest") {
		t.Errorf("images missing nginx:latest:\n%s", out)
	}
	if !strings.Contains(out, "111122223333") {
		t.Errorf("images missing short id:\n%s", out)
	}
	if !strings.Contains(out, "187.0MB") {
		t.Errorf("images missing human size:\n%s", out)
	}
	if !strings.Contains(out, "<none>") {
		t.Errorf("images missing dangling <none> row:\n%s", out)
	}
}

func TestLogsCommand(t *testing.T) {
	useFake(t, &fakedocker.Fake{
		Logs: map[string]string{"c1": "line one\nline two\n"},
	})
	out := run(t, "logs", "c1")
	if out != "line one\nline two\n" {
		t.Errorf("logs output = %q", out)
	}
}

func TestFormatHelpers(t *testing.T) {
	if got := shortID("sha256:0123456789abcdef"); got != "0123456789ab" {
		t.Errorf("shortID = %q", got)
	}
	if got := humanSize(999); got != "999B" {
		t.Errorf("humanSize(999) = %q", got)
	}
	if got := humanSize(1_500_000); got != "1.5MB" {
		t.Errorf("humanSize(1.5e6) = %q", got)
	}
	if got := truncate("abcdef", 4); got != "abc…" {
		t.Errorf("truncate = %q", got)
	}
	if repo, tag := splitRepoTag("registry.io/team/app:v1"); repo != "registry.io/team/app" || tag != "v1" {
		t.Errorf("splitRepoTag = (%q, %q)", repo, tag)
	}
	if repo, tag := splitRepoTag("busybox"); repo != "busybox" || tag != "<none>" {
		t.Errorf("splitRepoTag(bare) = (%q, %q)", repo, tag)
	}
}
