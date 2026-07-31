package fakedocker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
)

func TestFakeLifecycleChangesState(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", State: "exited"}}}
	ctx := context.Background()

	if err := f.StartContainer(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if f.Containers[0].State != "running" {
		t.Fatalf("state = %q, want running", f.Containers[0].State)
	}
	if err := f.StopContainer(ctx, "a", docker.StopOptions{}); err != nil {
		t.Fatal(err)
	}
	if f.Containers[0].State != "exited" {
		t.Fatalf("state = %q, want exited", f.Containers[0].State)
	}
	if err := f.StartContainer(ctx, "missing"); !errors.Is(err, docker.ErrNotFound) {
		t.Fatalf("missing container: got %v", err)
	}
}

func TestFakeRemoveRequiresConsent(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", State: "exited"}}}
	ctx := context.Background()

	if err := f.RemoveContainer(ctx, "a", docker.RemoveContainerOptions{}); !errors.Is(err, docker.ErrConsentRequired) {
		t.Fatalf("no consent: got %v, want ErrConsentRequired", err)
	}
	if len(f.Containers) != 1 {
		t.Fatal("container removed without consent")
	}
	if err := f.RemoveContainer(ctx, "a", docker.RemoveContainerOptions{Consent: docker.Confirm()}); err != nil {
		t.Fatalf("with consent: %v", err)
	}
	if len(f.Containers) != 0 {
		t.Fatal("container not removed with consent")
	}
}

func TestFakeRemoveRunningNeedsForce(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", State: "running"}}}
	ctx := context.Background()
	if err := f.RemoveContainer(ctx, "a", docker.RemoveContainerOptions{Consent: docker.Confirm()}); err == nil {
		t.Fatal("expected error removing running container without force")
	}
	if err := f.RemoveContainer(ctx, "a", docker.RemoveContainerOptions{Consent: docker.Confirm(), Force: true}); err != nil {
		t.Fatalf("force remove: %v", err)
	}
	if len(f.Containers) != 0 {
		t.Fatal("running container not force-removed")
	}
}

func TestFakePruneContainers(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{
		{ID: "run", State: "running"},
		{ID: "gone", State: "exited"},
		{ID: "dead", State: "dead"},
	}}
	ctx := context.Background()

	// Dry run needs no consent and removes nothing.
	rep, err := f.PruneContainers(ctx, docker.PruneOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.DryRun || len(rep.Items) != 2 {
		t.Fatalf("dry run report = %+v, want 2 items", rep)
	}
	if len(f.Containers) != 3 {
		t.Fatal("dry run removed containers")
	}

	// No consent => refused.
	if _, err := f.PruneContainers(ctx, docker.PruneOptions{}); !errors.Is(err, docker.ErrConsentRequired) {
		t.Fatalf("prune without consent: got %v", err)
	}

	// With consent => stopped/dead removed, running kept.
	rep, err = f.PruneContainers(ctx, docker.PruneOptions{Consent: docker.Confirm()})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Items) != 2 || len(f.Containers) != 1 || f.Containers[0].ID != "run" {
		t.Fatalf("after prune: report=%+v remaining=%+v", rep, f.Containers)
	}
}

func TestFakePruneImagesReclaimsSpace(t *testing.T) {
	f := &fakedocker.Fake{Images: []docker.ImageSummary{
		{ID: "dangling", RepoTags: nil, Size: 500},
		{ID: "tagged", RepoTags: []string{"nginx:latest"}, Size: 900},
	}}
	rep, err := f.PruneImages(context.Background(), docker.PruneOptions{Consent: docker.Confirm()})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Items) != 1 || rep.SpaceReclaimed != 500 {
		t.Fatalf("report = %+v, want 1 item / 500 bytes", rep)
	}
	if len(f.Images) != 1 || f.Images[0].ID != "tagged" {
		t.Fatalf("tagged image should remain: %+v", f.Images)
	}
}

func TestFakePruneVolumesDryRunUnsupported(t *testing.T) {
	f := &fakedocker.Fake{}
	if _, err := f.PruneVolumes(context.Background(), docker.PruneOptions{DryRun: true}); !errors.Is(err, docker.ErrDryRunUnsupported) {
		t.Fatalf("got %v, want ErrDryRunUnsupported", err)
	}
}

func TestFakeCreateNetwork(t *testing.T) {
	f := &fakedocker.Fake{}
	id, err := f.CreateNetwork(context.Background(), docker.CreateNetworkOptions{Name: "mynet"})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" || len(f.Networks) != 1 || f.Networks[0].Name != "mynet" || f.Networks[0].Driver != "bridge" {
		t.Fatalf("create network: id=%q nets=%+v", id, f.Networks)
	}
}
