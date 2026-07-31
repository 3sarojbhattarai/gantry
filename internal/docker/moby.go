package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/client"
)

// mobyClient implements Client on top of the official Docker SDK. Phase 0
// wires up construction, Ping, and Close so the interface is satisfied and the
// daemon connection is proven; the read paths return ErrNotImplemented until
// Phase 1.
type mobyClient struct {
	cli *client.Client
}

// compile-time assertion that mobyClient satisfies the interface.
var _ Client = (*mobyClient)(nil)

// New constructs a Client backed by the Docker SDK, configured from the
// environment (DOCKER_HOST, TLS settings, …) with API-version negotiation so
// it works against a range of daemon versions.
func New() (Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("gantry: creating docker client: %w", err)
	}
	return &mobyClient{cli: cli}, nil
}

func (m *mobyClient) Ping(ctx context.Context) error {
	if _, err := m.cli.Ping(ctx); err != nil {
		return fmt.Errorf("gantry: pinging docker daemon: %w", err)
	}
	return nil
}

func (m *mobyClient) ListContainers(ctx context.Context, all bool) ([]Container, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) InspectContainer(ctx context.Context, id string) (*ContainerDetails, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) ListImages(ctx context.Context) ([]ImageSummary, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) InspectImage(ctx context.Context, id string) (*ImageDetails, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) ListNetworks(ctx context.Context) ([]Network, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) ListVolumes(ctx context.Context) ([]Volume, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) ContainerLogs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) ContainerStats(ctx context.Context, id string) (<-chan Stats, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) Events(ctx context.Context) (<-chan Event, error) {
	return nil, ErrNotImplemented
}

func (m *mobyClient) Close() error {
	return m.cli.Close()
}
