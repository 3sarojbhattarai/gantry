//go:build integration

// Integration tests exercise the moby-backed engine against a real Docker
// daemon. They are excluded from the default build by the `integration` tag;
// run them with `make test-integration` (or `go test -tags integration`).
//
// Fixtures are created in a dedicated gantry-test-* namespace and torn down by
// removing anything matching that prefix, so a crashed run leaves nothing
// behind on the next invocation.
package docker

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"
)

const (
	testPrefix = "gantry-test-"
	testImage  = "alpine:3.20"
)

// setup returns a live engine client plus the raw SDK client used only to
// create fixtures (container create is not part of the engine until Phase 3).
func setup(t *testing.T) (Client, *dockerclient.Client, context.Context) {
	t.Helper()
	raw, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("raw client: %v", err)
	}
	ctx := context.Background()
	if _, err := raw.Ping(ctx); err != nil {
		t.Skipf("no reachable docker daemon: %v", err)
	}

	eng, err := New()
	if err != nil {
		t.Fatalf("engine client: %v", err)
	}
	t.Cleanup(func() { eng.Close(); raw.Close() })

	pullImage(t, raw, ctx)
	cleanupFixtures(t, raw, ctx) // clear anything a prior crashed run left
	t.Cleanup(func() { cleanupFixtures(t, raw, context.Background()) })
	return eng, raw, ctx
}

func pullImage(t *testing.T, raw *dockerclient.Client, ctx context.Context) {
	t.Helper()
	if _, err := raw.ImageInspect(ctx, testImage); err == nil {
		return
	}
	rc, err := raw.ImagePull(ctx, testImage, image.PullOptions{})
	if err != nil {
		t.Fatalf("pull %s: %v", testImage, err)
	}
	_, _ = io.Copy(io.Discard, rc)
	rc.Close()
}

// runFixture creates and starts a container named gantry-test-<suffix> that
// prints marker then sleeps, and returns its ID.
func runFixture(t *testing.T, raw *dockerclient.Client, ctx context.Context, suffix, marker string) string {
	t.Helper()
	name := testPrefix + suffix
	created, err := raw.ContainerCreate(ctx,
		&container.Config{Image: testImage, Cmd: []string{"sh", "-c", "echo " + marker + "; sleep 300"}},
		&container.HostConfig{}, nil, nil, name)
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	if err := raw.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		t.Fatalf("start %s: %v", name, err)
	}
	return created.ID
}

func cleanupFixtures(t *testing.T, raw *dockerclient.Client, ctx context.Context) {
	t.Helper()
	list, err := raw.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", testPrefix)),
	})
	if err != nil {
		return
	}
	for _, c := range list {
		_ = raw.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
	}
}

func TestIntegrationListAndInspectContainer(t *testing.T) {
	eng, raw, ctx := setup(t)
	id := runFixture(t, raw, ctx, "ps", "hello")

	cs, err := eng.ListContainers(ctx, true)
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	var found bool
	for _, c := range cs {
		if c.ID == id {
			found = true
			if len(c.Names) == 0 || !strings.HasPrefix(c.Names[0], testPrefix) {
				t.Errorf("unexpected names: %v", c.Names)
			}
		}
	}
	if !found {
		t.Fatalf("fixture %s not in ListContainers", id[:12])
	}

	d, err := eng.InspectContainer(ctx, id)
	if err != nil {
		t.Fatalf("InspectContainer: %v", err)
	}
	if d.State != "running" {
		t.Errorf("state = %q, want running", d.State)
	}
}

func TestIntegrationContainerLogsDemux(t *testing.T) {
	eng, raw, ctx := setup(t)
	id := runFixture(t, raw, ctx, "logs", "marker-line")

	// Give the container a moment to emit its line.
	time.Sleep(500 * time.Millisecond)

	rc, err := eng.ContainerLogs(ctx, id, LogOptions{Tail: "all"})
	if err != nil {
		t.Fatalf("ContainerLogs: %v", err)
	}
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	if !strings.Contains(string(b), "marker-line") {
		t.Fatalf("logs = %q, want to contain marker-line", b)
	}
}

func TestIntegrationInspectNotFound(t *testing.T) {
	eng, _, ctx := setup(t)
	_, err := eng.InspectContainer(ctx, "gantry-test-does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing container")
	}
}

func TestIntegrationStopAndRemove(t *testing.T) {
	eng, raw, ctx := setup(t)
	id := runFixture(t, raw, ctx, "mutate", "hi")

	// Removing without consent must be refused, and the container must survive.
	if err := eng.RemoveContainer(ctx, id, RemoveContainerOptions{Force: true}); err != ErrConsentRequired {
		t.Fatalf("remove without consent: got %v, want ErrConsentRequired", err)
	}

	if err := eng.StopContainer(ctx, id, StopOptions{}); err != nil {
		t.Fatalf("StopContainer: %v", err)
	}
	d, err := eng.InspectContainer(ctx, id)
	if err != nil {
		t.Fatalf("InspectContainer: %v", err)
	}
	if d.State != "exited" {
		t.Errorf("state after stop = %q, want exited", d.State)
	}

	// With consent, removal succeeds and the container disappears.
	if err := eng.RemoveContainer(ctx, id, RemoveContainerOptions{Consent: Confirm()}); err != nil {
		t.Fatalf("RemoveContainer with consent: %v", err)
	}
	cs, err := eng.ListContainers(ctx, true)
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	for _, c := range cs {
		if c.ID == id {
			t.Fatal("container still present after removal")
		}
	}
}

func TestIntegrationListImages(t *testing.T) {
	eng, _, ctx := setup(t)
	imgs, err := eng.ListImages(ctx)
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	// The pulled fixture image must be present.
	var found bool
	for _, im := range imgs {
		for _, rt := range im.RepoTags {
			if rt == testImage {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("fixture image %s not listed", testImage)
	}
}
