package fakedocker

import (
	"context"
	"errors"

	"github.com/3sarojbhattarai/gantry/internal/docker"
)

// --- lifecycle --------------------------------------------------------------

func (f *Fake) StartContainer(_ context.Context, id string) error {
	return f.setState(id, "running")
}

func (f *Fake) StopContainer(_ context.Context, id string, _ docker.StopOptions) error {
	return f.setState(id, "exited")
}

func (f *Fake) RestartContainer(_ context.Context, id string, _ docker.StopOptions) error {
	return f.setState(id, "running")
}

func (f *Fake) KillContainer(_ context.Context, id, _ string) error {
	return f.setState(id, "exited")
}

func (f *Fake) setState(id, state string) error {
	for i := range f.Containers {
		if f.Containers[i].ID == id {
			f.Containers[i].State = state
			return nil
		}
	}
	return docker.ErrNotFound
}

// --- removals ---------------------------------------------------------------

func (f *Fake) RemoveContainer(_ context.Context, id string, opts docker.RemoveContainerOptions) error {
	if !opts.Granted() {
		return docker.ErrConsentRequired
	}
	for i, c := range f.Containers {
		if c.ID == id {
			if c.State == "running" && !opts.Force {
				return errors.New("fakedocker: container is running (needs force)")
			}
			f.Containers = append(f.Containers[:i], f.Containers[i+1:]...)
			return nil
		}
	}
	return docker.ErrNotFound
}

func (f *Fake) RemoveImage(_ context.Context, id string, opts docker.RemoveImageOptions) error {
	if !opts.Granted() {
		return docker.ErrConsentRequired
	}
	for i, im := range f.Images {
		if im.ID == id {
			f.Images = append(f.Images[:i], f.Images[i+1:]...)
			return nil
		}
	}
	return docker.ErrNotFound
}

func (f *Fake) RemoveNetwork(_ context.Context, id string, opts docker.RemoveNetworkOptions) error {
	if !opts.Granted() {
		return docker.ErrConsentRequired
	}
	for i, n := range f.Networks {
		if n.ID == id || n.Name == id {
			f.Networks = append(f.Networks[:i], f.Networks[i+1:]...)
			return nil
		}
	}
	return docker.ErrNotFound
}

func (f *Fake) RemoveVolume(_ context.Context, name string, opts docker.RemoveVolumeOptions) error {
	if !opts.Granted() {
		return docker.ErrConsentRequired
	}
	for i, v := range f.Volumes {
		if v.Name == name {
			f.Volumes = append(f.Volumes[:i], f.Volumes[i+1:]...)
			return nil
		}
	}
	return docker.ErrNotFound
}

// --- prune ------------------------------------------------------------------

func (f *Fake) PruneContainers(_ context.Context, opts docker.PruneOptions) (docker.PruneReport, error) {
	if opts.DryRun {
		var items []string
		for _, c := range f.Containers {
			if fakePrunable(c.State) {
				items = append(items, c.ID)
			}
		}
		return docker.PruneReport{Items: items, DryRun: true}, nil
	}
	if !opts.Granted() {
		return docker.PruneReport{}, docker.ErrConsentRequired
	}
	var kept []docker.Container
	var removed []string
	for _, c := range f.Containers {
		if fakePrunable(c.State) {
			removed = append(removed, c.ID)
		} else {
			kept = append(kept, c)
		}
	}
	f.Containers = kept
	return docker.PruneReport{Items: removed}, nil
}

func (f *Fake) PruneImages(_ context.Context, opts docker.PruneOptions) (docker.PruneReport, error) {
	collect := func() ([]string, uint64) {
		var items []string
		var space uint64
		for _, im := range f.Images {
			if fakeDangling(im) {
				items = append(items, im.ID)
				if im.Size > 0 {
					space += uint64(im.Size)
				}
			}
		}
		return items, space
	}
	if opts.DryRun {
		items, space := collect()
		return docker.PruneReport{Items: items, SpaceReclaimed: space, DryRun: true}, nil
	}
	if !opts.Granted() {
		return docker.PruneReport{}, docker.ErrConsentRequired
	}
	items, space := collect()
	var kept []docker.ImageSummary
	for _, im := range f.Images {
		if !fakeDangling(im) {
			kept = append(kept, im)
		}
	}
	f.Images = kept
	return docker.PruneReport{Items: items, SpaceReclaimed: space}, nil
}

func (f *Fake) PruneVolumes(_ context.Context, opts docker.PruneOptions) (docker.PruneReport, error) {
	if opts.DryRun {
		return docker.PruneReport{DryRun: true}, docker.ErrDryRunUnsupported
	}
	if !opts.Granted() {
		return docker.PruneReport{}, docker.ErrConsentRequired
	}
	removed := make([]string, 0, len(f.Volumes))
	for _, v := range f.Volumes {
		removed = append(removed, v.Name)
	}
	f.Volumes = nil
	return docker.PruneReport{Items: removed}, nil
}

func (f *Fake) PruneNetworks(_ context.Context, opts docker.PruneOptions) (docker.PruneReport, error) {
	if opts.DryRun {
		return docker.PruneReport{DryRun: true}, docker.ErrDryRunUnsupported
	}
	if !opts.Granted() {
		return docker.PruneReport{}, docker.ErrConsentRequired
	}
	var kept []docker.Network
	var removed []string
	for _, n := range f.Networks {
		switch n.Name {
		case "bridge", "host", "none":
			kept = append(kept, n)
		default:
			removed = append(removed, n.Name)
		}
	}
	f.Networks = kept
	return docker.PruneReport{Items: removed}, nil
}

// --- networks ---------------------------------------------------------------

func (f *Fake) CreateNetwork(_ context.Context, opts docker.CreateNetworkOptions) (string, error) {
	driver := opts.Driver
	if driver == "" {
		driver = "bridge"
	}
	id := "net-" + opts.Name
	f.Networks = append(f.Networks, docker.Network{
		ID:       id,
		Name:     opts.Name,
		Driver:   driver,
		Scope:    "local",
		Internal: opts.Internal,
		Labels:   opts.Labels,
	})
	return id, nil
}

func (f *Fake) ConnectNetwork(_ context.Context, _, _ string) error            { return nil }
func (f *Fake) DisconnectNetwork(_ context.Context, _, _ string, _ bool) error { return nil }

// --- helpers ----------------------------------------------------------------

func fakePrunable(state string) bool {
	switch state {
	case "exited", "created", "dead":
		return true
	default:
		return false
	}
}

func fakeDangling(im docker.ImageSummary) bool {
	if len(im.RepoTags) == 0 {
		return true
	}
	for _, t := range im.RepoTags {
		if t == "<none>:<none>" {
			return true
		}
	}
	return false
}
