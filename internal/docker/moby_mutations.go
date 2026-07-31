package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
)

// --- lifecycle (reversible; no consent needed) ------------------------------

func (m *mobyClient) StartContainer(ctx context.Context, id string) error {
	if err := m.cli.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return fmt.Errorf("gantry: starting container %s: %w", id, mapErr(err))
	}
	return nil
}

func (m *mobyClient) StopContainer(ctx context.Context, id string, opts StopOptions) error {
	if err := m.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: opts.Timeout}); err != nil {
		return fmt.Errorf("gantry: stopping container %s: %w", id, mapErr(err))
	}
	return nil
}

func (m *mobyClient) RestartContainer(ctx context.Context, id string, opts StopOptions) error {
	if err := m.cli.ContainerRestart(ctx, id, container.StopOptions{Timeout: opts.Timeout}); err != nil {
		return fmt.Errorf("gantry: restarting container %s: %w", id, mapErr(err))
	}
	return nil
}

func (m *mobyClient) KillContainer(ctx context.Context, id, signal string) error {
	if err := m.cli.ContainerKill(ctx, id, signal); err != nil {
		return fmt.Errorf("gantry: killing container %s: %w", id, mapErr(err))
	}
	return nil
}

// --- destructive removals (consent-gated) -----------------------------------

func (m *mobyClient) RemoveContainer(ctx context.Context, id string, opts RemoveContainerOptions) error {
	if !opts.Granted() {
		return ErrConsentRequired
	}
	err := m.cli.ContainerRemove(ctx, id, container.RemoveOptions{
		Force:         opts.Force,
		RemoveVolumes: opts.RemoveVolumes,
	})
	if err != nil {
		return fmt.Errorf("gantry: removing container %s: %w", id, mapErr(err))
	}
	return nil
}

func (m *mobyClient) RemoveImage(ctx context.Context, id string, opts RemoveImageOptions) error {
	if !opts.Granted() {
		return ErrConsentRequired
	}
	_, err := m.cli.ImageRemove(ctx, id, image.RemoveOptions{
		Force:         opts.Force,
		PruneChildren: opts.PruneChildren,
	})
	if err != nil {
		return fmt.Errorf("gantry: removing image %s: %w", id, mapErr(err))
	}
	return nil
}

func (m *mobyClient) RemoveNetwork(ctx context.Context, id string, opts RemoveNetworkOptions) error {
	if !opts.Granted() {
		return ErrConsentRequired
	}
	if err := m.cli.NetworkRemove(ctx, id); err != nil {
		return fmt.Errorf("gantry: removing network %s: %w", id, mapErr(err))
	}
	return nil
}

func (m *mobyClient) RemoveVolume(ctx context.Context, name string, opts RemoveVolumeOptions) error {
	if !opts.Granted() {
		return ErrConsentRequired
	}
	if err := m.cli.VolumeRemove(ctx, name, opts.Force); err != nil {
		return fmt.Errorf("gantry: removing volume %s: %w", name, mapErr(err))
	}
	return nil
}

// --- prune (consent-gated unless DryRun) ------------------------------------

func (m *mobyClient) PruneContainers(ctx context.Context, opts PruneOptions) (PruneReport, error) {
	if opts.DryRun {
		return previewContainerPrune(ctx, m)
	}
	if !opts.Granted() {
		return PruneReport{}, ErrConsentRequired
	}
	r, err := m.cli.ContainersPrune(ctx, filters.NewArgs())
	if err != nil {
		return PruneReport{}, fmt.Errorf("gantry: pruning containers: %w", mapErr(err))
	}
	return PruneReport{Items: r.ContainersDeleted, SpaceReclaimed: r.SpaceReclaimed}, nil
}

func (m *mobyClient) PruneImages(ctx context.Context, opts PruneOptions) (PruneReport, error) {
	if opts.DryRun {
		return previewImagePrune(ctx, m)
	}
	if !opts.Granted() {
		return PruneReport{}, ErrConsentRequired
	}
	r, err := m.cli.ImagesPrune(ctx, filters.NewArgs())
	if err != nil {
		return PruneReport{}, fmt.Errorf("gantry: pruning images: %w", mapErr(err))
	}
	items := make([]string, 0, len(r.ImagesDeleted))
	for _, d := range r.ImagesDeleted {
		switch {
		case d.Deleted != "":
			items = append(items, "deleted "+d.Deleted)
		case d.Untagged != "":
			items = append(items, "untagged "+d.Untagged)
		}
	}
	return PruneReport{Items: items, SpaceReclaimed: r.SpaceReclaimed}, nil
}

func (m *mobyClient) PruneVolumes(ctx context.Context, opts PruneOptions) (PruneReport, error) {
	if opts.DryRun {
		return PruneReport{DryRun: true}, ErrDryRunUnsupported
	}
	if !opts.Granted() {
		return PruneReport{}, ErrConsentRequired
	}
	r, err := m.cli.VolumesPrune(ctx, filters.NewArgs())
	if err != nil {
		return PruneReport{}, fmt.Errorf("gantry: pruning volumes: %w", mapErr(err))
	}
	return PruneReport{Items: r.VolumesDeleted, SpaceReclaimed: r.SpaceReclaimed}, nil
}

func (m *mobyClient) PruneNetworks(ctx context.Context, opts PruneOptions) (PruneReport, error) {
	if opts.DryRun {
		return PruneReport{DryRun: true}, ErrDryRunUnsupported
	}
	if !opts.Granted() {
		return PruneReport{}, ErrConsentRequired
	}
	r, err := m.cli.NetworksPrune(ctx, filters.NewArgs())
	if err != nil {
		return PruneReport{}, fmt.Errorf("gantry: pruning networks: %w", mapErr(err))
	}
	return PruneReport{Items: r.NetworksDeleted}, nil
}

// --- networks ---------------------------------------------------------------

func (m *mobyClient) CreateNetwork(ctx context.Context, opts CreateNetworkOptions) (string, error) {
	driver := opts.Driver
	if driver == "" {
		driver = "bridge"
	}
	resp, err := m.cli.NetworkCreate(ctx, opts.Name, network.CreateOptions{
		Driver:   driver,
		Internal: opts.Internal,
		Labels:   opts.Labels,
	})
	if err != nil {
		return "", fmt.Errorf("gantry: creating network %s: %w", opts.Name, mapErr(err))
	}
	return resp.ID, nil
}

func (m *mobyClient) ConnectNetwork(ctx context.Context, networkID, containerID string) error {
	if err := m.cli.NetworkConnect(ctx, networkID, containerID, nil); err != nil {
		return fmt.Errorf("gantry: connecting %s to network %s: %w", containerID, networkID, mapErr(err))
	}
	return nil
}

func (m *mobyClient) DisconnectNetwork(ctx context.Context, networkID, containerID string, force bool) error {
	if err := m.cli.NetworkDisconnect(ctx, networkID, containerID, force); err != nil {
		return fmt.Errorf("gantry: disconnecting %s from network %s: %w", containerID, networkID, mapErr(err))
	}
	return nil
}

// --- shared dry-run previews (interface-based, so consistent everywhere) -----

// previewContainerPrune reports the stopped containers a prune would remove.
// Container disk usage is not part of the list view, so SpaceReclaimed is left
// zero for the preview.
func previewContainerPrune(ctx context.Context, c Client) (PruneReport, error) {
	cs, err := c.ListContainers(ctx, true)
	if err != nil {
		return PruneReport{}, err
	}
	var items []string
	for _, x := range cs {
		if prunableContainerState(x.State) {
			items = append(items, x.ID)
		}
	}
	return PruneReport{Items: items, DryRun: true}, nil
}

// previewImagePrune reports the dangling images a prune would remove, summing
// their sizes.
func previewImagePrune(ctx context.Context, c Client) (PruneReport, error) {
	imgs, err := c.ListImages(ctx)
	if err != nil {
		return PruneReport{}, err
	}
	var (
		items []string
		space uint64
	)
	for _, im := range imgs {
		if danglingImage(im) {
			items = append(items, im.ID)
			if im.Size > 0 {
				space += uint64(im.Size)
			}
		}
	}
	return PruneReport{Items: items, SpaceReclaimed: space, DryRun: true}, nil
}

func prunableContainerState(state string) bool {
	switch state {
	case "exited", "created", "dead":
		return true
	default:
		return false
	}
}

func danglingImage(im ImageSummary) bool {
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
