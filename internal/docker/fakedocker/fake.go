// Package fakedocker provides an in-memory implementation of docker.Client for
// use in tests. Engine and CLI table tests run against it, so they need no
// running daemon. Populate the exported slices with fixtures and, when a test
// needs to exercise an error path, set the matching *Err field.
package fakedocker

import (
	"context"
	"io"
	"strings"

	"github.com/3sarojbhattarai/gantry/internal/docker"
)

// Fake is an in-memory docker.Client. The zero value is usable and reports an
// empty daemon. Fields are read directly by the List/Inspect methods; the *Err
// fields let a test force any method to fail.
type Fake struct {
	Containers []docker.Container
	Images     []docker.ImageSummary
	Networks   []docker.Network
	Volumes    []docker.Volume

	// Stream fixtures replayed by the streaming methods, in order.
	StatsSamples []docker.Stats
	EventStream  []docker.Event

	// Logs returned by ContainerLogs, keyed by container ID.
	Logs map[string]string

	// Error injection. When non-nil, the corresponding method returns the
	// error instead of its normal result. PingErr also stands in for a
	// generally unreachable daemon.
	PingErr             error
	ListContainersErr   error
	InspectContainerErr error
	ListImagesErr       error
	InspectImageErr     error
	ListNetworksErr     error
	ListVolumesErr      error
	ContainerLogsErr    error
	ContainerStatsErr   error
	EventsErr           error
}

var _ docker.Client = (*Fake)(nil)

func (f *Fake) Ping(ctx context.Context) error { return f.PingErr }

func (f *Fake) ListContainers(ctx context.Context, all bool) ([]docker.Container, error) {
	if f.ListContainersErr != nil {
		return nil, f.ListContainersErr
	}
	if all {
		return f.Containers, nil
	}
	var running []docker.Container
	for _, c := range f.Containers {
		if c.State == "running" {
			running = append(running, c)
		}
	}
	return running, nil
}

func (f *Fake) InspectContainer(ctx context.Context, id string) (*docker.ContainerDetails, error) {
	if f.InspectContainerErr != nil {
		return nil, f.InspectContainerErr
	}
	for _, c := range f.Containers {
		if c.ID == id {
			return &docker.ContainerDetails{Container: c}, nil
		}
	}
	return nil, docker.ErrNotFound
}

func (f *Fake) ListImages(ctx context.Context) ([]docker.ImageSummary, error) {
	if f.ListImagesErr != nil {
		return nil, f.ListImagesErr
	}
	return f.Images, nil
}

func (f *Fake) InspectImage(ctx context.Context, id string) (*docker.ImageDetails, error) {
	if f.InspectImageErr != nil {
		return nil, f.InspectImageErr
	}
	for _, img := range f.Images {
		if img.ID == id {
			return &docker.ImageDetails{ImageSummary: img}, nil
		}
	}
	return nil, docker.ErrNotFound
}

func (f *Fake) ListNetworks(ctx context.Context) ([]docker.Network, error) {
	if f.ListNetworksErr != nil {
		return nil, f.ListNetworksErr
	}
	return f.Networks, nil
}

func (f *Fake) ListVolumes(ctx context.Context) ([]docker.Volume, error) {
	if f.ListVolumesErr != nil {
		return nil, f.ListVolumesErr
	}
	return f.Volumes, nil
}

func (f *Fake) ContainerLogs(ctx context.Context, id string, opts docker.LogOptions) (io.ReadCloser, error) {
	if f.ContainerLogsErr != nil {
		return nil, f.ContainerLogsErr
	}
	return io.NopCloser(strings.NewReader(f.Logs[id])), nil
}

func (f *Fake) ContainerStats(ctx context.Context, id string) (<-chan docker.Stats, error) {
	if f.ContainerStatsErr != nil {
		return nil, f.ContainerStatsErr
	}
	ch := make(chan docker.Stats)
	go func() {
		defer close(ch)
		for _, s := range f.StatsSamples {
			select {
			case <-ctx.Done():
				return
			case ch <- s:
			}
		}
	}()
	return ch, nil
}

func (f *Fake) Events(ctx context.Context) (<-chan docker.Event, error) {
	if f.EventsErr != nil {
		return nil, f.EventsErr
	}
	ch := make(chan docker.Event)
	go func() {
		defer close(ch)
		for _, e := range f.EventStream {
			select {
			case <-ctx.Done():
				return
			case ch <- e:
			}
		}
	}()
	return ch, nil
}

func (f *Fake) Close() error { return nil }
