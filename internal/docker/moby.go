package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// mobyClient implements Client on top of the official Docker SDK. This file is
// the one and only place moby's types are allowed to appear: every method
// translates them into gantry's domain types before returning, so nothing
// above the engine ever imports github.com/docker/docker.
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
	summaries, err := m.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, fmt.Errorf("gantry: listing containers: %w", mapErr(err))
	}
	out := make([]Container, 0, len(summaries))
	for _, s := range summaries {
		out = append(out, toContainer(s))
	}
	return out, nil
}

func (m *mobyClient) InspectContainer(ctx context.Context, id string) (*ContainerDetails, error) {
	r, err := m.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("gantry: inspecting container %s: %w", id, mapErr(err))
	}
	return toContainerDetails(r), nil
}

func (m *mobyClient) ListImages(ctx context.Context) ([]ImageSummary, error) {
	imgs, err := m.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("gantry: listing images: %w", mapErr(err))
	}
	out := make([]ImageSummary, 0, len(imgs))
	for _, im := range imgs {
		out = append(out, toImageSummary(im))
	}
	return out, nil
}

func (m *mobyClient) InspectImage(ctx context.Context, id string) (*ImageDetails, error) {
	r, err := m.cli.ImageInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("gantry: inspecting image %s: %w", id, mapErr(err))
	}
	return &ImageDetails{
		ImageSummary: ImageSummary{
			ID:       r.ID,
			RepoTags: r.RepoTags,
			Created:  parseTime(r.Created),
			Size:     r.Size,
		},
		RepoDigests:  r.RepoDigests,
		Architecture: r.Architecture,
		Os:           r.Os,
		Author:       r.Author,
	}, nil
}

func (m *mobyClient) ListNetworks(ctx context.Context) ([]Network, error) {
	nets, err := m.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("gantry: listing networks: %w", mapErr(err))
	}
	out := make([]Network, 0, len(nets))
	for _, n := range nets {
		out = append(out, toNetwork(n))
	}
	return out, nil
}

func (m *mobyClient) ListVolumes(ctx context.Context) ([]Volume, error) {
	resp, err := m.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("gantry: listing volumes: %w", mapErr(err))
	}
	out := make([]Volume, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		if v == nil {
			continue
		}
		out = append(out, toVolume(v))
	}
	return out, nil
}

func (m *mobyClient) ContainerLogs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error) {
	// Whether the stream is multiplexed depends on the container's TTY setting,
	// which only inspect reveals. A TTY stream is raw; a non-TTY stream needs
	// demuxing.
	insp, err := m.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("gantry: inspecting container %s for logs: %w", id, mapErr(err))
	}

	tail := opts.Tail
	if tail == "" {
		tail = "all"
	}
	rc, err := m.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     opts.Follow,
		Timestamps: opts.Timestamps,
		Tail:       tail,
	})
	if err != nil {
		return nil, fmt.Errorf("gantry: reading container %s logs: %w", id, mapErr(err))
	}

	if insp.Config != nil && insp.Config.Tty {
		return rc, nil // raw, not multiplexed
	}
	return newDemuxReader(rc), nil
}

func (m *mobyClient) ContainerStats(ctx context.Context, id string) (<-chan Stats, error) {
	resp, err := m.cli.ContainerStats(ctx, id, true)
	if err != nil {
		return nil, fmt.Errorf("gantry: streaming container %s stats: %w", id, mapErr(err))
	}
	ch := make(chan Stats)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)
		for {
			var raw container.StatsResponse
			if err := dec.Decode(&raw); err != nil {
				return // EOF, or ctx cancellation surfaced as a read error
			}
			select {
			case <-ctx.Done():
				return
			case ch <- computeStats(id, raw):
			}
		}
	}()
	return ch, nil
}

func (m *mobyClient) Events(ctx context.Context) (<-chan Event, error) {
	msgCh, errCh := m.cli.Events(ctx, events.ListOptions{})
	out := make(chan Event)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-errCh:
				return // stream error or clean shutdown; either way we're done
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case out <- toEvent(msg):
				}
			}
		}
	}()
	return out, nil
}

func (m *mobyClient) Close() error {
	return m.cli.Close()
}

// --- moby -> domain translation (the boundary lives here) -------------------

func toContainer(s container.Summary) Container {
	return Container{
		ID:      s.ID,
		Names:   trimNames(s.Names),
		Image:   s.Image,
		ImageID: s.ImageID,
		Command: s.Command,
		Created: time.Unix(s.Created, 0),
		State:   string(s.State),
		Status:  s.Status,
		Ports:   toPorts(s.Ports),
		Labels:  s.Labels,
	}
}

func toContainerDetails(r container.InspectResponse) *ContainerDetails {
	d := &ContainerDetails{}
	if b := r.ContainerJSONBase; b != nil {
		d.ID = b.ID
		d.Names = []string{strings.TrimPrefix(b.Name, "/")}
		d.ImageID = b.Image
		d.Created = parseTime(b.Created)
		d.Command = strings.TrimSpace(b.Path + " " + strings.Join(b.Args, " "))
		d.Path = b.Path
		d.Args = b.Args
		d.Platform = b.Platform
		d.RestartCount = b.RestartCount
		if b.State != nil {
			d.State = string(b.State.Status)
			d.ExitCode = b.State.ExitCode
			d.Error = b.State.Error
			d.StartedAt = parseTime(b.State.StartedAt)
			d.FinishedAt = parseTime(b.State.FinishedAt)
		}
	}
	if r.Config != nil {
		d.Image = r.Config.Image
		d.Env = r.Config.Env
		d.Labels = r.Config.Labels
	}
	return d
}

func toImageSummary(im image.Summary) ImageSummary {
	return ImageSummary{
		ID:       im.ID,
		RepoTags: im.RepoTags,
		Created:  time.Unix(im.Created, 0),
		Size:     im.Size,
		Labels:   im.Labels,
	}
}

func toNetwork(n network.Summary) Network {
	return Network{
		ID:       n.ID,
		Name:     n.Name,
		Driver:   n.Driver,
		Scope:    n.Scope,
		Created:  n.Created,
		Internal: n.Internal,
		Labels:   n.Labels,
	}
}

func toVolume(v *volume.Volume) Volume {
	return Volume{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		Created:    parseTime(v.CreatedAt),
		Labels:     v.Labels,
	}
}

func toEvent(m events.Message) Event {
	return Event{
		Type:   string(m.Type),
		Action: string(m.Action),
		Actor:  m.Actor.ID,
		Name:   m.Actor.Attributes["name"],
		Time:   time.Unix(0, m.TimeNano),
	}
}

func toPorts(ps []container.Port) []Port {
	if len(ps) == 0 {
		return nil
	}
	out := make([]Port, 0, len(ps))
	for _, p := range ps {
		out = append(out, Port{IP: p.IP, Private: p.PrivatePort, Public: p.PublicPort, Type: p.Type})
	}
	return out
}

// computeStats reduces a raw stats sample into gantry's Stats, applying the
// same CPU and memory formulas the `docker stats` CLI uses. It takes moby's
// type directly so the math is unit-testable without a daemon.
func computeStats(id string, s container.StatsResponse) Stats {
	st := Stats{ContainerID: id}

	// CPU percentage: the container's CPU-time delta as a fraction of the
	// system's CPU-time delta, scaled by the number of online CPUs.
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemUsage) - float64(s.PreCPUStats.SystemUsage)
	online := float64(s.CPUStats.OnlineCPUs)
	if online == 0 {
		online = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
	}
	if sysDelta > 0 && cpuDelta > 0 {
		st.CPUPercent = (cpuDelta / sysDelta) * online * 100.0
	}

	// Memory: exclude page cache so the figure matches `docker stats`. The
	// cache key differs between cgroup v1 (total_inactive_file) and v2
	// (inactive_file).
	usage := s.MemoryStats.Usage
	if v, ok := s.MemoryStats.Stats["total_inactive_file"]; ok && v < usage {
		usage -= v
	} else if v, ok := s.MemoryStats.Stats["inactive_file"]; ok && v < usage {
		usage -= v
	}
	st.MemUsage = usage
	st.MemLimit = s.MemoryStats.Limit
	if s.MemoryStats.Limit > 0 {
		st.MemPercent = float64(usage) / float64(s.MemoryStats.Limit) * 100.0
	}

	for _, n := range s.Networks {
		st.NetRx += n.RxBytes
		st.NetTx += n.TxBytes
	}
	for _, e := range s.BlkioStats.IoServiceBytesRecursive {
		switch strings.ToLower(e.Op) {
		case "read":
			st.BlockRead += e.Value
		case "write":
			st.BlockWrite += e.Value
		}
	}
	st.PIDs = s.PidsStats.Current
	return st
}

// --- small helpers ----------------------------------------------------------

// mapErr converts SDK sentinel errors into gantry's own so callers can branch
// on them with errors.Is without importing the SDK.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if client.IsErrNotFound(err) {
		return ErrNotFound
	}
	return err
}

// parseTime parses an RFC3339(nano) timestamp, returning the zero time on any
// error or an empty string.
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// trimNames strips the leading slash the daemon puts on container names.
func trimNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		out = append(out, strings.TrimPrefix(n, "/"))
	}
	return out
}
