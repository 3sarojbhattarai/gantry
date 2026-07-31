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

	// Close releases any resources held by the client (e.g. the SDK's HTTP
	// connections). It is safe to call on a fake.
	Close() error
}
