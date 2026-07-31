package docker

import (
	"context"
	"io"
)

// Client is gantry's read-only view of the Docker daemon. Every renderer (CLI,
// TUI, API) depends on this interface rather than any concrete implementation,
// so the moby-backed client and the in-memory fake are interchangeable.
//
// Mutations (start/stop/remove/prune/…) arrive in Phase 3 as an extension of
// this interface; keeping the read surface separate keeps the read paths — and
// their tests — honest.
type Client interface {
	// Ping verifies the daemon is reachable.
	Ping(ctx context.Context) error

	ListContainers(ctx context.Context, all bool) ([]Container, error)
	InspectContainer(ctx context.Context, id string) (*ContainerDetails, error)

	ListImages(ctx context.Context) ([]ImageSummary, error)
	InspectImage(ctx context.Context, id string) (*ImageDetails, error)

	ListNetworks(ctx context.Context) ([]Network, error)
	ListVolumes(ctx context.Context) ([]Volume, error)

	// ContainerLogs returns already-demuxed, human-readable output. The caller
	// is responsible for closing the returned reader.
	ContainerLogs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error)

	// ContainerStats streams resource-usage samples for one container. The
	// returned channel is closed when ctx is cancelled or the stream ends.
	ContainerStats(ctx context.Context, id string) (<-chan Stats, error)

	// Events streams daemon events. The returned channel is closed when ctx is
	// cancelled or the stream ends.
	Events(ctx context.Context) (<-chan Event, error)

	// --- mutations (Phase 3) ---
	//
	// Lifecycle operations are reversible and need no consent. Removals and
	// prunes are destructive: they refuse to run unless their options carry
	// granted Consent, returning ErrConsentRequired otherwise.

	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string, opts StopOptions) error
	RestartContainer(ctx context.Context, id string, opts StopOptions) error
	KillContainer(ctx context.Context, id, signal string) error

	RemoveContainer(ctx context.Context, id string, opts RemoveContainerOptions) error
	RemoveImage(ctx context.Context, id string, opts RemoveImageOptions) error
	RemoveNetwork(ctx context.Context, id string, opts RemoveNetworkOptions) error
	RemoveVolume(ctx context.Context, name string, opts RemoveVolumeOptions) error

	PruneContainers(ctx context.Context, opts PruneOptions) (PruneReport, error)
	PruneImages(ctx context.Context, opts PruneOptions) (PruneReport, error)
	PruneVolumes(ctx context.Context, opts PruneOptions) (PruneReport, error)
	PruneNetworks(ctx context.Context, opts PruneOptions) (PruneReport, error)

	CreateNetwork(ctx context.Context, opts CreateNetworkOptions) (string, error)
	ConnectNetwork(ctx context.Context, networkID, containerID string) error
	DisconnectNetwork(ctx context.Context, networkID, containerID string, force bool) error

	// --- exec + create (Phases 6-7) ---

	// ContainerExec starts a command inside a running container and returns a
	// live bidirectional session.
	ContainerExec(ctx context.Context, id string, opts ExecOptions) (ExecSession, error)

	// CreateContainer creates a container from spec, optionally starting it, and
	// returns its ID.
	CreateContainer(ctx context.Context, spec CreateSpec, start bool) (string, error)

	// SpecFromContainer reads an existing container's config into a CreateSpec
	// (the --from "clone and tweak" path).
	SpecFromContainer(ctx context.Context, id string) (CreateSpec, error)

	// Close releases any resources held by the client (e.g. the SDK's HTTP
	// connections). It is safe to call on a fake.
	Close() error
}
