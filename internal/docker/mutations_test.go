package docker

import (
	"context"
	"errors"
	"testing"
)

func TestConsentZeroValueNotGranted(t *testing.T) {
	var c Consent
	if c.Granted() {
		t.Fatal("zero Consent must not be granted")
	}
	if !Confirm().Granted() {
		t.Fatal("Confirm() must be granted")
	}
}

// The consent gate short-circuits before any SDK call, so we can verify it
// against a real (unconnected) client with no daemon running.
func TestDestructiveOpsRequireConsent(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()
	ctx := context.Background()

	checks := map[string]error{
		"RemoveContainer": c.RemoveContainer(ctx, "x", RemoveContainerOptions{}),
		"RemoveImage":     c.RemoveImage(ctx, "x", RemoveImageOptions{}),
		"RemoveNetwork":   c.RemoveNetwork(ctx, "x", RemoveNetworkOptions{}),
		"RemoveVolume":    c.RemoveVolume(ctx, "x", RemoveVolumeOptions{}),
	}
	for name, got := range checks {
		if !errors.Is(got, ErrConsentRequired) {
			t.Errorf("%s without consent: got %v, want ErrConsentRequired", name, got)
		}
	}

	if _, err := c.PruneContainers(ctx, PruneOptions{}); !errors.Is(err, ErrConsentRequired) {
		t.Errorf("PruneContainers without consent: got %v", err)
	}
}

func TestPruneDryRunUnsupportedForVolumesAndNetworks(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()
	ctx := context.Background()

	if _, err := c.PruneVolumes(ctx, PruneOptions{DryRun: true}); !errors.Is(err, ErrDryRunUnsupported) {
		t.Errorf("PruneVolumes dry-run: got %v, want ErrDryRunUnsupported", err)
	}
	if _, err := c.PruneNetworks(ctx, PruneOptions{DryRun: true}); !errors.Is(err, ErrDryRunUnsupported) {
		t.Errorf("PruneNetworks dry-run: got %v, want ErrDryRunUnsupported", err)
	}
}

func TestPrunableContainerState(t *testing.T) {
	for _, s := range []string{"exited", "created", "dead"} {
		if !prunableContainerState(s) {
			t.Errorf("%q should be prunable", s)
		}
	}
	for _, s := range []string{"running", "paused", "restarting"} {
		if prunableContainerState(s) {
			t.Errorf("%q should not be prunable", s)
		}
	}
}

func TestDanglingImage(t *testing.T) {
	if !danglingImage(ImageSummary{RepoTags: nil}) {
		t.Error("no tags => dangling")
	}
	if !danglingImage(ImageSummary{RepoTags: []string{"<none>:<none>"}}) {
		t.Error("<none>:<none> => dangling")
	}
	if danglingImage(ImageSummary{RepoTags: []string{"nginx:latest"}}) {
		t.Error("tagged => not dangling")
	}
}
